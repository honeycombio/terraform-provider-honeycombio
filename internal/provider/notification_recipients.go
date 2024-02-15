package provider

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/exp/slices"

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
					},
				},
			},
		},
	}
}

func reconcileNotificationRecipientState(remote []client.NotificationRecipient, state []models.NotificationRecipientModel) []models.NotificationRecipientModel {
	if state == nil {
		// if we don't have any state, we can't reconcile anything so just return the remote recipients
		return flattenNotificationRecipients(remote)
	}

	recipients := make([]models.NotificationRecipientModel, len(remote))
	// match the remote recipients to those in the state sorting out type+target vs ID
	for i, r := range remote {
		idx := slices.IndexFunc(state, func(s models.NotificationRecipientModel) bool {
			if !s.ID.IsNull() {
				return s.ID.ValueString() == r.ID
			}
			return s.Type.ValueString() == string(r.Type) && s.Target.ValueString() == r.Target
		})
		if idx < 0 {
			// if we didn't find a match, use the recipient as specified in remote
			recipients[i] = notificationRecipientToModel(r)
		} else {
			// if we found a match, use the stored recipient
			recipients[i] = state[idx]
		}
	}

	return recipients
}

func expandNotificationRecipients(n []models.NotificationRecipientModel) []client.NotificationRecipient {
	recipients := make([]client.NotificationRecipient, len(n))

	for i, r := range n {
		rcpt := client.NotificationRecipient{
			ID:     r.ID.ValueString(),
			Type:   client.RecipientType(r.Type.ValueString()),
			Target: r.Target.ValueString(),
		}
		if r.Details != nil {
			rcpt.Details = &client.NotificationRecipientDetails{
				PDSeverity: client.PagerDutySeverity(r.Details[0].PDSeverity.ValueString()),
			}
		}
		recipients[i] = rcpt
	}

	return recipients
}

func flattenNotificationRecipients(n []client.NotificationRecipient) []models.NotificationRecipientModel {
	recipients := make([]models.NotificationRecipientModel, len(n))
	for i, r := range n {
		recipients[i] = notificationRecipientToModel(r)
	}

	return recipients
}

func notificationRecipientToModel(r client.NotificationRecipient) models.NotificationRecipientModel {
	rcpt := models.NotificationRecipientModel{
		ID:     types.StringValue(r.ID),
		Type:   types.StringValue(string(r.Type)),
		Target: types.StringValue(r.Target),
	}
	if r.Details != nil {
		rcpt.Details = []models.NotificationRecipientDetailsModel{
			{PDSeverity: types.StringValue(string(r.Details.PDSeverity))},
		}
	}

	return rcpt
}
