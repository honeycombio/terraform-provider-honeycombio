package modifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/honeycombio/terraform-provider-honeycombio/internal/models"
)

type triggerNotificationRecipientsModifier struct{}

var _ planmodifier.Set = &triggerNotificationRecipientsModifier{}

func (m triggerNotificationRecipientsModifier) Description(_ context.Context) string {
	return "Handles the default states and value manipulations for a Trigger's notification recipients."
}

func (m triggerNotificationRecipientsModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

// PlanModifySet is the honeycombio_trigger counterpart of notificationRecipientsModifier.
// It exists as a separate modifier only because a Trigger's recipient carries the two
// per-group routing attributes, so it must decode into TriggerNotificationRecipientModel --
// decoding a six-attribute object into the four-field base model is a runtime error.
//
// The per-group routing attributes need no handling here. They are Optional and never
// Computed, so they are never unknown in a plan and pass through untouched.
func (m triggerNotificationRecipientsModifier) PlanModifySet(ctx context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
	// Do nothing on resource destroy.
	if req.Plan.Raw.IsNull() {
		return
	}
	// Nothing to reconcile against on create.
	if req.StateValue.IsNull() {
		return
	}
	// An unknown set (a dynamic block over an unknown for_each, say) has no elements to
	// normalize, and decoding one into the model is an error.
	if req.PlanValue.IsUnknown() {
		return
	}

	var rcpts []models.TriggerNotificationRecipientModel
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("recipient"), &rcpts)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// manage null values properly based on the type of recipient
	for i := range rcpts {
		normalizeRecipientIdentity(&rcpts[i].NotificationRecipientModel)
	}

	updated, diag := types.SetValueFrom(ctx, req.PlanValue.ElementType(ctx), rcpts)
	resp.Diagnostics.Append(diag...)
	resp.PlanValue = updated
}

func TriggerNotificationRecipients() planmodifier.Set {
	return triggerNotificationRecipientsModifier{}
}
