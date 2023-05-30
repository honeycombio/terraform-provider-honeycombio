package modifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/models"
)

type notificationRecipientsModifier struct{}

var _ planmodifier.Set = &notificationRecipientsModifier{}

func (m notificationRecipientsModifier) Description(_ context.Context) string {
	return "Handles the default states and value manipulations for notificiation recipients."
}

func (m notificationRecipientsModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m notificationRecipientsModifier) PlanModifySet(ctx context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
	// Do nothing on resource destroy.
	if req.Plan.Raw.IsNull() {
		return
	}

	var plan models.TriggerResourceModel
	req.Plan.Get(ctx, &plan)

	var rcpts []models.NotificationRecipientModel
	req.Plan.GetAttribute(ctx, path.Root("recipient"), &rcpts)

	// manage null values properly based on the type of recipient
	if !req.StateValue.IsNull() {
		for i, r := range rcpts {
			if r.ID.IsUnknown() && r.Type.IsUnknown() {
				// likely dependant on creation of another resource
				continue
			}
			if r.ID.IsUnknown() && !r.Type.IsUnknown() {
				// specified by type and target
				r.ID = types.StringNull()
				if r.Type.ValueString() == string(client.RecipientTypePagerDuty) {
					// PagerDuty recipients do not have a target
					r.Target = types.StringNull()
				}
			}
			if !r.ID.IsUnknown() && r.Type.IsUnknown() {
				// specified by ID
				r.Type = types.StringNull()
				r.Target = types.StringNull()
			}

			rcpts[i] = r
		}

		updated, diag := types.SetValueFrom(ctx, req.PlanValue.ElementType(ctx), rcpts)
		resp.Diagnostics.Append(diag...)
		resp.PlanValue = updated
	}
}

func NotificationRecipients() planmodifier.Set {
	return notificationRecipientsModifier{}
}
