package provider

import (
	"context"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/validation"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/modifiers"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/models"
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
						Blocks: map[string]schema.Block{
							"variable": schema.SetNestedBlock{
								Description: "The variables to set with the webhook notification.",
								Validators: []validator.Set{
									setvalidator.SizeAtMost(10),
									validation.ValidNotificationVariables(),
								},
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"name": schema.StringAttribute{
											Required:    true,
											Description: "The name of the variable",
											Validators: []validator.String{
												stringvalidator.LengthBetween(1, 64),
												stringvalidator.RegexMatches(webhookTemplateNameRegex, "must be an alphanumeric string beginning with a lowercase letter"),
											},
										},
										"value": schema.StringAttribute{
											Description: "The value of the variable",
											Optional:    true,
											Computed:    true,
											Default:     stringdefault.StaticString(""),
											Validators: []validator.String{
												stringvalidator.LengthAtMost(256),
											},
										},
									},
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
	if len(remote) == 0 || state.IsUnknown() {
		return types.SetNull(types.ObjectType{AttrTypes: models.NotificationRecipientAttrType})
	}

	if state.IsNull() {
		// if we don't have any state, we can't reconcile anything so
		// just return the remote recipients biased toward id over type+target
		var values []attr.Value
		for _, r := range remote {
			recipObj, d := types.ObjectValue(models.NotificationRecipientAttrType, map[string]attr.Value{
				"id":                   types.StringValue(r.ID),
				"type":                 types.StringNull(),
				"target":               types.StringNull(),
				"notification_details": types.ListNull(types.ObjectType{AttrTypes: models.NotificationRecipientDetailsAttrType}),
			})
			diags.Append(d...)

			values = append(values, recipObj)
		}
		result, d := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: models.NotificationRecipientAttrType}, values)
		diags.Append(d...)

		return result
	}

	var recipients []models.NotificationRecipientModel
	diags.Append(state.ElementsAs(ctx, &recipients, false)...)
	if diags.HasError() {
		return types.SetNull(types.ObjectType{AttrTypes: models.NotificationRecipientAttrType})
	}
	mappedRecips := mapNotificationRecipientToState(ctx, remote, recipients, diags)

	var values []attr.Value
	for _, r := range mappedRecips {
		values = append(values, notificationRecipientModelToObjectValue(ctx, r, diags))
	}
	result, d := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: models.NotificationRecipientAttrType}, values)
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
				return nil
			}

			rcpt.Details = &client.NotificationRecipientDetails{
				PDSeverity: client.PagerDutySeverity(details[0].PDSeverity.ValueString()),
				Variables:  expandNotificationVariables(ctx, details[0].Variables, diags),
			}
		}
		clientRecips[i] = rcpt
	}

	return clientRecips
}
func notificationRecipientModelToObjectValue(ctx context.Context, r models.NotificationRecipientModel, diags *diag.Diagnostics) basetypes.ObjectValue {
	recipObj := map[string]attr.Value{
		"id":     r.ID,
		"type":   r.Type,
		"target": r.Target,
	}

	var result basetypes.ListValue
	if !r.Details.IsNull() && !r.Details.IsUnknown() {
		var details []models.NotificationRecipientDetailsModel
		diags.Append(r.Details.ElementsAs(ctx, &details, false)...)
		if diags.HasError() {
			return basetypes.ObjectValue{}
		}

		detailsObjVal, d := types.ObjectValue(models.NotificationRecipientDetailsAttrType, map[string]attr.Value{"pagerduty_severity": details[0].PDSeverity, "variable": details[0].Variables})
		diags.Append(d...)
		result, d = types.ListValueFrom(ctx, types.ObjectType{AttrTypes: models.NotificationRecipientDetailsAttrType}, []attr.Value{detailsObjVal})
		diags.Append(d...)
	} else {
		result = types.ListNull(types.ObjectType{AttrTypes: models.NotificationRecipientDetailsAttrType})
	}

	recipObj["notification_details"] = result
	recipObjVal, d := types.ObjectValue(models.NotificationRecipientAttrType, recipObj)
	diags.Append(d...)

	return recipObjVal
}

func notificationRecipientToModel(ctx context.Context, r client.NotificationRecipient, diags *diag.Diagnostics) models.NotificationRecipientModel {
	rcpt := models.NotificationRecipientModel{
		ID:      types.StringValue(r.ID),
		Type:    types.StringValue(string(r.Type)),
		Target:  types.StringValue(r.Target),
		Details: notificationRecipientDetailsToList(ctx, r.Details, diags),
	}

	return rcpt
}

func notificationRecipientDetailsToList(ctx context.Context, details *client.NotificationRecipientDetails, diags *diag.Diagnostics) basetypes.ListValue {
	var result basetypes.ListValue
	if details != nil {
		detailsObj := map[string]attr.Value{"pagerduty_severity": types.StringValue(string(details.PDSeverity)), "variable": flattenNotificationVariables(ctx, details.Variables, diags)}
		objVal, d := types.ObjectValue(models.NotificationRecipientDetailsAttrType, detailsObj)
		diags.Append(d...)
		result, d = types.ListValueFrom(ctx, types.ObjectType{AttrTypes: models.NotificationRecipientDetailsAttrType}, []attr.Value{objVal})
		diags.Append(d...)
	} else {
		result = types.ListNull(types.ObjectType{AttrTypes: models.NotificationRecipientDetailsAttrType})
	}

	return result
}

func notificationVariableToObjectValue(v client.NotificationVariable, diags *diag.Diagnostics) basetypes.ObjectValue {
	variableObj := map[string]attr.Value{
		"name":  types.StringValue(v.Name),
		"value": types.StringValue(v.Value),
	}
	varObjVal, d := types.ObjectValue(models.NotificationVariableAttrType, variableObj)
	diags.Append(d...)

	return varObjVal
}

func flattenNotificationVariables(ctx context.Context, vars []client.NotificationVariable, diags *diag.Diagnostics) types.Set {
	if len(vars) == 0 || vars == nil {
		return types.SetNull(types.ObjectType{AttrTypes: models.NotificationVariableAttrType})
	}

	var notifVarValues []attr.Value
	for _, v := range vars {
		notifVarValues = append(notifVarValues, notificationVariableToObjectValue(v, diags))
	}
	notifVarResult, d := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: models.WebhookHeaderAttrType}, notifVarValues)
	diags.Append(d...)

	return notifVarResult
}

func expandNotificationVariables(ctx context.Context, set types.Set, diags *diag.Diagnostics) []client.NotificationVariable {
	var notifVars []models.NotificationVariableModel
	diags.Append(set.ElementsAs(ctx, &notifVars, false)...)
	if diags.HasError() {
		return nil
	}

	clientNotifVars := make([]client.NotificationVariable, len(notifVars))
	for i, v := range notifVars {
		notifVar := client.NotificationVariable{
			Name:  v.Name.ValueString(),
			Value: v.Value.ValueString(),
		}

		clientNotifVars[i] = notifVar
	}

	return clientNotifVars
}
