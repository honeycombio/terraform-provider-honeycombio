package provider

import (
	"context"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/models"
)

// triggerRecipientRoutingAttributes are the per-group routing attributes added to
// honeycombio_trigger's `recipient` block. They are not part of the block shared with
// honeycombio_burn_alert: the Burn Alerts API rejects them, and Burn Alerts have no query
// groups to route.
//
// Both are Optional and deliberately NOT Computed. The Trigger resource writes config
// straight back to state after apply (see the Create and Update implementations), so a
// Computed attribute whose config went null would have its prior value carried into the plan
// while state took null -- Terraform then cannot correlate the planned set element with the
// applied one. Optional-only keeps plan, config and state identical by construction, and
// means the attributes can actually be removed from HCL.
func triggerRecipientRoutingAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"group_filter": schema.MapAttribute{
			Optional:    true,
			ElementType: models.GroupFilterValueType,
			Description: "Only notify this recipient about the query groups matching this filter. " +
				"Maps a group by column of the Trigger's query to the values which route to this recipient. " +
				"Omit for a catch-all recipient which is notified about every group. " +
				"Requires an `alert_type` of `on_group_change` and a query with at least one group by. " +
				"Only one routing rule is allowed per recipient.",
			Validators: []validator.Map{
				mapvalidator.SizeAtLeast(1),
				mapvalidator.KeysAre(stringvalidator.LengthAtLeast(1)),
				mapvalidator.ValueSetsAre(
					setvalidator.SizeAtLeast(1),
					setvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
				),
			},
		},
		"pagerduty_per_group_incidents": schema.BoolAttribute{
			Optional: true,
			Description: "Open and resolve one PagerDuty incident per triggered group instead of one " +
				"incident per Trigger. Only supported for PagerDuty recipients, and requires an " +
				"`alert_type` of `on_group_change` and a query with at least one group by.",
		},
	}
}

// validateGroupedRecipientRouting enforces, at plan time, the rules the API would otherwise
// reject with a 422.
//
// q is the Trigger's query spec, or nil when the Trigger references a query by ID -- in
// which case the group by columns are unknown to us and the two rules which need them are
// skipped, leaving those to the API.
//
// Two rules can only be checked partially:
//   - a recipient given by `id` has an unknown type, so we cannot tell whether it is a
//     PagerDuty recipient.
//   - `group_filter` column names cannot be checked for a query_id Trigger.
//
// The team entitlement for grouped resolution alerts is not detectable from the provider,
// so that is left to the API error.
func validateGroupedRecipientRouting(
	ctx context.Context,
	data models.TriggerResourceModel,
	q *client.QuerySpec,
	resp *resource.ValidateConfigResponse,
) {
	if data.Recipients.IsNull() || data.Recipients.IsUnknown() {
		return
	}

	var rcpts []models.TriggerNotificationRecipientModel
	resp.Diagnostics.Append(data.Recipients.ElementsAs(ctx, &rcpts, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// alert_type is Optional and Computed with a static default, so it is null in config
	// when unset. Treat that as the default rather than as "no opinion".
	alertType := client.TriggerAlertTypeOnChange
	if !data.AlertType.IsNull() && !data.AlertType.IsUnknown() {
		alertType = client.TriggerAlertType(data.AlertType.ValueString())
	}
	groupedAlerts := alertType == client.TriggerAlertTypeOnGroupChange

	seenIDs := make(map[string]int, len(rcpts))
	seenTargets := make(map[string]int, len(rcpts))

	for i, rcpt := range rcpts {
		recipPath := path.Root("recipient").AtListIndex(i)
		hasGroupFilter := !rcpt.GroupFilter.IsNull() && !rcpt.GroupFilter.IsUnknown()
		// An unknown value still means the argument was configured.
		hasPerGroupIncidents := !rcpt.PDPerGroupIncidents.IsNull()

		// per-group routing requires the grouped resolution alert type
		if (hasGroupFilter || hasPerGroupIncidents) && !groupedAlerts {
			attr := "group_filter"
			if !hasGroupFilter {
				attr = "pagerduty_per_group_incidents"
			}
			resp.Diagnostics.AddAttributeError(
				recipPath.AtName(attr),
				"Trigger validation error",
				`Per-group recipient routing requires an "alert_type" of "on_group_change". `+
					"Honeycomb only evaluates per-group state for that alert type, and will "+
					"otherwise discard the routing configured here.",
			)
		}

		// pagerduty_per_group_incidents is only supported for PagerDuty recipients.
		// A recipient given by ID has an unknown type, so it can only be checked when the
		// type was configured explicitly.
		if hasPerGroupIncidents && !rcpt.Type.IsNull() && !rcpt.Type.IsUnknown() &&
			rcpt.Type.ValueString() != string(client.RecipientTypePagerDuty) {
			resp.Diagnostics.AddAttributeError(
				recipPath.AtName("pagerduty_per_group_incidents"),
				"Trigger validation error",
				`"pagerduty_per_group_incidents" is only supported for PagerDuty recipients, `+
					"but this recipient is of type "+rcpt.Type.ValueString()+".",
			)
		}

		// only one routing rule is allowed per recipient
		if !rcpt.ID.IsNull() && !rcpt.ID.IsUnknown() {
			id := rcpt.ID.ValueString()
			if _, ok := seenIDs[id]; ok {
				resp.Diagnostics.AddAttributeError(
					recipPath.AtName("id"),
					"Conflicting configuration arguments",
					"Only one routing rule is allowed per recipient, but recipient "+id+
						" is also configured earlier in this Trigger. Combine the rules into a "+
						"single \"recipient\" block: a group_filter can list several values, and "+
						"several columns.",
				)
			}
			seenIDs[id] = i
		} else if !rcpt.Type.IsNull() && !rcpt.Type.IsUnknown() &&
			!rcpt.Target.IsUnknown() {
			key := rcpt.Type.ValueString() + "\x00" + rcpt.Target.ValueString()
			if _, ok := seenTargets[key]; ok {
				resp.Diagnostics.AddAttributeError(
					recipPath.AtName("target"),
					"Conflicting configuration arguments",
					"Only one routing rule is allowed per recipient, but this "+
						rcpt.Type.ValueString()+" recipient is also configured earlier in this "+
						"Trigger. Combine the rules into a single \"recipient\" block.",
				)
			}
			seenTargets[key] = i
		}

		if q == nil || !hasGroupFilter {
			continue
		}

		// on_group_change routing needs something to group by
		if len(q.Breakdowns) == 0 {
			resp.Diagnostics.AddAttributeError(
				recipPath.AtName("group_filter"),
				"Trigger validation error",
				`Per-group recipient routing requires the Trigger's query to group by at `+
					"least one column, but the query has no breakdowns.",
			)
			continue
		}

		// a group_filter can only name columns the query groups by
		var filter map[string][]string
		resp.Diagnostics.Append(rcpt.GroupFilter.ElementsAs(ctx, &filter, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		for column := range filter {
			if !slices.Contains(q.Breakdowns, column) {
				resp.Diagnostics.AddAttributeError(
					recipPath.AtName("group_filter"),
					"Trigger validation error",
					`"`+column+`" is not one of the Trigger query's group by columns `+
						"("+strings.Join(q.Breakdowns, ", ")+"). A group_filter can only route "+
						"on a column the query groups by.",
				)
			}
		}
	}
}

// expandTriggerNotificationRecipients is the honeycombio_trigger counterpart of
// expandNotificationRecipients, adding the per-group routing fields.
func expandTriggerNotificationRecipients(ctx context.Context, set types.Set, diags *diag.Diagnostics) []client.TriggerNotificationRecipient {
	var recipients []models.TriggerNotificationRecipientModel
	diags.Append(set.ElementsAs(ctx, &recipients, false)...)
	if diags.HasError() {
		return nil
	}

	clientRecips := make([]client.TriggerNotificationRecipient, len(recipients))
	for i, r := range recipients {
		clientRecips[i] = client.TriggerNotificationRecipient{
			NotificationRecipient: expandNotificationRecipient(ctx, r.NotificationRecipientModel, diags),
			GroupFilter:           expandGroupFilter(ctx, r.GroupFilter, diags),
			PDPerGroupIncidents:   r.PDPerGroupIncidents.ValueBoolPointer(),
		}
		if diags.HasError() {
			return nil
		}
	}

	return clientRecips
}

func mapTriggerNotificationRecipientToState(
	ctx context.Context,
	remote []client.TriggerNotificationRecipient,
	state []models.TriggerNotificationRecipientModel,
	diags *diag.Diagnostics,
) []models.TriggerNotificationRecipientModel {
	recipients := make([]models.TriggerNotificationRecipientModel, len(remote))
	// match the remote recipients to those in the state
	// in an effort to preserve the id vs type+target distinction
	for i, r := range remote {
		idx := slices.IndexFunc(state, func(s models.TriggerNotificationRecipientModel) bool {
			if !s.ID.IsNull() {
				return s.ID.ValueString() == r.ID
			}
			return s.Type.ValueString() == string(r.Type) && s.Target.ValueString() == r.Target
		})
		if idx < 0 {
			// if we didn't find a match, use the recipient as specified in remote
			recipients[i] = triggerNotificationRecipientToModel(ctx, r, diags)
		} else {
			// if we found a match, use the stored recipient
			recipients[i] = state[idx]
		}
	}
	return recipients
}

// reconcileReadTriggerNotificationRecipientState is the honeycombio_trigger counterpart of
// reconcileReadNotificationRecipientState.
func reconcileReadTriggerNotificationRecipientState(
	ctx context.Context,
	remote []client.TriggerNotificationRecipient,
	state types.Set,
	diags *diag.Diagnostics,
) types.Set {
	recipientType := types.ObjectType{AttrTypes: models.TriggerNotificationRecipientAttrType}

	if len(remote) == 0 || state.IsUnknown() {
		return types.SetNull(recipientType)
	}

	if state.IsNull() {
		// if we don't have any state, we can't reconcile anything so
		// just return the remote recipients biased toward id over type+target
		//
		// The routing attributes are nulled here for the same reason type, target and
		// notification_details are: on import we cannot tell how the recipient was
		// authored, so the first plan after an import shows the routing as configured.
		var values []attr.Value
		for _, r := range remote {
			recipObj, d := types.ObjectValue(models.TriggerNotificationRecipientAttrType, map[string]attr.Value{
				"id":                            types.StringValue(r.ID),
				"type":                          types.StringNull(),
				"target":                        types.StringNull(),
				"notification_details":          types.ListNull(types.ObjectType{AttrTypes: models.NotificationRecipientDetailsAttrType}),
				"group_filter":                  types.MapNull(models.GroupFilterValueType),
				"pagerduty_per_group_incidents": types.BoolNull(),
			})
			diags.Append(d...)

			values = append(values, recipObj)
		}
		result, d := types.SetValueFrom(ctx, recipientType, values)
		diags.Append(d...)

		return result
	}

	var recipients []models.TriggerNotificationRecipientModel
	diags.Append(state.ElementsAs(ctx, &recipients, false)...)
	if diags.HasError() {
		return types.SetNull(recipientType)
	}
	mappedRecips := mapTriggerNotificationRecipientToState(ctx, remote, recipients, diags)

	var values []attr.Value
	for _, r := range mappedRecips {
		values = append(values, triggerNotificationRecipientModelToObjectValue(ctx, r, diags))
	}
	result, d := types.SetValueFrom(ctx, recipientType, values)
	diags.Append(d...)

	return result
}

func triggerNotificationRecipientModelToObjectValue(
	ctx context.Context,
	r models.TriggerNotificationRecipientModel,
	diags *diag.Diagnostics,
) basetypes.ObjectValue {
	base := notificationRecipientModelToObjectValue(ctx, r.NotificationRecipientModel, diags)
	if diags.HasError() {
		return basetypes.ObjectValue{}
	}

	// Attributes() returns a copy, so this does not mutate the base object.
	recipObj := base.Attributes()

	// A zero-valued types.Map is null but carries a nil element type, which will not
	// type-check against the object's Map(Set(String)) attribute. Models reach this
	// function straight from state, so normalize before building the object.
	groupFilter := r.GroupFilter
	if groupFilter.IsNull() || groupFilter.IsUnknown() {
		groupFilter = types.MapNull(models.GroupFilterValueType)
	}
	recipObj["group_filter"] = groupFilter
	recipObj["pagerduty_per_group_incidents"] = r.PDPerGroupIncidents

	recipObjVal, d := types.ObjectValue(models.TriggerNotificationRecipientAttrType, recipObj)
	diags.Append(d...)

	return recipObjVal
}

func triggerNotificationRecipientToModel(
	ctx context.Context,
	r client.TriggerNotificationRecipient,
	diags *diag.Diagnostics,
) models.TriggerNotificationRecipientModel {
	return models.TriggerNotificationRecipientModel{
		NotificationRecipientModel: notificationRecipientToModel(ctx, r.NotificationRecipient, diags),
		GroupFilter:                flattenGroupFilter(ctx, r.GroupFilter, diags),
		PDPerGroupIncidents:        flattenPerGroupIncidents(r.PDPerGroupIncidents),
	}
}

// flattenGroupFilter normalizes an absent and an empty filter to null: the API omits
// group_filter entirely for a catch-all recipient, which is what an omitted attribute means.
func flattenGroupFilter(ctx context.Context, filter map[string][]string, diags *diag.Diagnostics) types.Map {
	if len(filter) == 0 {
		return types.MapNull(models.GroupFilterValueType)
	}

	result, d := types.MapValueFrom(ctx, models.GroupFilterValueType, filter)
	diags.Append(d...)

	return result
}

func expandGroupFilter(ctx context.Context, filter types.Map, diags *diag.Diagnostics) map[string][]string {
	if filter.IsNull() || filter.IsUnknown() {
		return nil
	}

	var result map[string][]string
	diags.Append(filter.ElementsAs(ctx, &result, false)...)
	if diags.HasError() {
		return nil
	}

	return result
}

// flattenPerGroupIncidents normalizes a false to null. The API returns
// pagerduty_per_group_incidents for every PagerDuty recipient -- false when it isn't in use
// -- and omits it for every other recipient type. Since false is indistinguishable from
// unset, only true becomes a known value, so an unmatched recipient read back from the API
// does not manufacture a diff against a config which omits the attribute.
func flattenPerGroupIncidents(perGroup *bool) types.Bool {
	if perGroup == nil || !*perGroup {
		return types.BoolNull()
	}

	return types.BoolValue(true)
}
