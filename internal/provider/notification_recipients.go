package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/modifiers"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/models"
	"golang.org/x/exp/slices"
)

func notificationRecipientSchema(allowedTypes []client.RecipientType) schema.SetNestedBlock {
	return schema.SetNestedBlock{
		Description:   "Zero or more recipients to notify when the resource fires.",
		PlanModifiers: []planmodifier.Set{modifiers.NotificationRecipients()},
		NestedObject: schema.NestedBlockObject{
			Validators: []validator.Object{
				objectvalidator.AtLeastOneOf(
					path.MatchRelative().AtName("id"),
					path.MatchRelative().AtName("type"),
				),
			},
			Attributes: map[string]schema.Attribute{
				"id": schema.StringAttribute{
					Optional:    true,
					Computed:    true,
					Description: "The ID of an existing recipient.",
					Validators: []validator.String{
						stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("type")),
						stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("target")),
					},
				},
				"type": schema.StringAttribute{
					Optional:    true,
					Computed:    true,
					Description: "The type of the notification recipient.",
					Validators: []validator.String{
						stringvalidator.OneOf(helper.AsStringSlice(allowedTypes)...),
					},
				},
				"target": schema.StringAttribute{
					Optional:    true,
					Computed:    true,
					Description: "Target of the notification, this has another meaning depending on the type of recipient.",
					Validators: []validator.String{
						stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("type")),
					},
				},
			},
			Blocks: map[string]schema.Block{
				"notification_details": schema.ListNestedBlock{
					Description: "Additional details to send along with the notification.",
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"pagerduty_severity": schema.StringAttribute{
								Optional:      true,
								Computed:      true,
								Description:   "The severity to set with the PagerDuty notification. If no severity is provided, 'critical' is assumed.",
								PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
								Validators: []validator.String{
									stringvalidator.All(
										stringvalidator.OneOf("info", "warning", "error", "critical"),
									),
								},
							},
						},
					},
				},
			},
		},
	}
}

func mapNotificationRecipientToState(ctx context.Context, remote []client.NotificationRecipient, state []models.NotificationRecipientModel, diags *diag.Diagnostics) []models.NotificationRecipientModel {
	recipients := make([]models.NotificationRecipientModel, len(remote))
	// match the remote recipients to those in the state
	// in an effort to preserve the id vs type+target distinction
	for i, r := range remote {
		idx := slices.IndexFunc(state, func(s models.NotificationRecipientModel) bool {
			if !s.ID.IsNull() {
				return s.ID.ValueString() == r.ID
			}
			return s.Type.ValueString() == string(r.Type) && s.Target.ValueString() == r.Target
		})
		if idx < 0 {
			// if we didn't find a match, use the recipient as specified in remote
			recipients[i] = notificationRecipientToModel(ctx, r, diags)
		} else {
			// if we found a match, use the stored recipient
			recipients[i] = state[idx]
		}
	}
	return recipients
}

func reconcileReadNotificationRecipientState(ctx context.Context, remote []client.NotificationRecipient, state types.Set, diags *diag.Diagnostics) types.Set {
	if state.IsNull() || state.IsUnknown() {
		// if we don't have any state, we can't reconcile anything so just return the remote recipients
		return flattenNotificationRecipients(ctx, remote, diags)
	}

	var recipients []models.NotificationRecipientModel
	diags.Append(state.ElementsAs(ctx, &recipients, false)...)
	if diags.HasError() {
		// do something
	}
	mappedRecips := mapNotificationRecipientToState(ctx, remote, recipients, diags)

	var values []attr.Value
	for _, r := range mappedRecips {
		values = append(values, NotificationRecipientModelToObjectValue(ctx, r, diags))
	}
	result, d := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: models.NotificationRecipientAttrTypes}, values)
	diags.Append(d...)

	return result
}

func expandNotificationRecipients(ctx context.Context, set types.Set, diags *diag.Diagnostics) []client.NotificationRecipient {
	var recipients []models.NotificationRecipientModel
	diags.Append(set.ElementsAs(ctx, &recipients, false)...)
	if diags.HasError() {
		return nil
	}

	clientRecips := make([]client.NotificationRecipient, len(recipients))
	for i, r := range recipients {
		rcpt := client.NotificationRecipient{
			ID:     r.ID.ValueString(),
			Type:   client.RecipientType(r.Type.ValueString()),
			Target: r.Target.ValueString(),
		}
		if !r.Details.IsNull() && !r.Details.IsUnknown() {
			var details []models.NotificationRecipientDetailsModel
			diags.Append(r.Details.ElementsAs(ctx, &details, false)...)
			if diags.HasError() {
				// do something
			}
			rcpt.Details = &client.NotificationRecipientDetails{
				PDSeverity: client.PagerDutySeverity(details[0].PDSeverity.ValueString()),
			}
		}
		clientRecips[i] = rcpt
	}

	return clientRecips
}

func flattenNotificationRecipients(ctx context.Context, n []client.NotificationRecipient, diags *diag.Diagnostics) types.Set {
	elemType := types.ObjectType{AttrTypes: models.NotificationRecipientAttrTypes}
	if len(n) == 0 {
		return types.SetValueMust(elemType, []attr.Value{})
	}

	var values []attr.Value
	for _, r := range n {
		values = append(values, notificationRecipientToObjectValue(ctx, r, diags))
	}
	result, d := types.SetValueFrom(ctx, elemType, values)
	diags.Append(d...)

	return result
}

func notificationRecipientToObjectValue(ctx context.Context, r client.NotificationRecipient, diags *diag.Diagnostics) basetypes.ObjectValue {
	recipObj := map[string]attr.Value{
		"id":     types.StringValue(r.ID),
		"type":   types.StringValue(string(r.Type)),
		"target": types.StringValue(r.Target),
	}

	var detailsObjVal basetypes.ObjectValue
	if r.Details != nil {
		detailsObj := map[string]attr.Value{"pagerduty_severity": types.StringValue(string(r.Details.PDSeverity))}
		var d diag.Diagnostics
		detailsObjVal, d = types.ObjectValue(models.NotificationRecipientDetailsAttrTypes, detailsObj)
		diags.Append(d...)
	} else {
		detailsObjVal = types.ObjectNull(models.NotificationRecipientDetailsAttrTypes)
	}
	result, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: models.NotificationRecipientDetailsAttrTypes}, []attr.Value{detailsObjVal})
	diags.Append(d...)
	recipObj["notification_details"] = result
	recipObjVal, d := types.ObjectValue(models.NotificationRecipientAttrTypes, recipObj)
	diags.Append(d...)

	return recipObjVal
}

func NotificationRecipientModelToObjectValue(ctx context.Context, r models.NotificationRecipientModel, diags *diag.Diagnostics) basetypes.ObjectValue {
	recipObj := map[string]attr.Value{
		"id":     r.ID,
		"type":   r.Type,
		"target": r.Target,
	}

	var detailsObjVal basetypes.ObjectValue
	if !r.Details.IsNull() && !r.Details.IsUnknown() {
		var details []models.NotificationRecipientDetailsModel
		diags.Append(r.Details.ElementsAs(ctx, &details, false)...)
		if diags.HasError() {
			return basetypes.ObjectValue{}
		}
		var d diag.Diagnostics
		detailsObjVal, d = types.ObjectValue(models.NotificationRecipientDetailsAttrTypes, map[string]attr.Value{"pagerduty_severity": details[0].PDSeverity})
		diags.Append(d...)
	} else {
		detailsObjVal = types.ObjectNull(models.NotificationRecipientDetailsAttrTypes)
	}
	result, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: models.NotificationRecipientDetailsAttrTypes}, []attr.Value{detailsObjVal})
	diags.Append(d...)
	recipObj["notification_details"] = result
	recipObjVal, d := types.ObjectValue(models.NotificationRecipientAttrTypes, recipObj)
	diags.Append(d...)

	return recipObjVal
}

func notificationRecipientToModel(ctx context.Context, r client.NotificationRecipient, diags *diag.Diagnostics) models.NotificationRecipientModel {
	rcpt := models.NotificationRecipientModel{
		ID:     types.StringValue(r.ID),
		Type:   types.StringValue(string(r.Type)),
		Target: types.StringValue(r.Target),
	}
	if r.Details != nil {
		detailsObj := map[string]attr.Value{"pagerduty_severity": types.StringValue(string(r.Details.PDSeverity))}
		objVal, d := types.ObjectValue(models.NotificationRecipientDetailsAttrTypes, detailsObj)
		diags.Append(d...)
		result, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: models.NotificationRecipientDetailsAttrTypes}, []attr.Value{objVal})
		diags.Append(d...)
		rcpt.Details = result
	}

	return rcpt
}
