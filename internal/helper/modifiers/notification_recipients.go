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
	// Nothing to reconcile against on create.
	if req.StateValue.IsNull() {
		return
	}

	var rcpts []models.NotificationRecipientModel
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("recipient"), &rcpts)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// manage null values properly based on the type of recipient
	for i := range rcpts {
		normalizeRecipientIdentity(&rcpts[i])
	}

	updated, diag := types.SetValueFrom(ctx, req.PlanValue.ElementType(ctx), rcpts)
	resp.Diagnostics.Append(diag...)
	resp.PlanValue = updated
}

// normalizeRecipientIdentity resolves the "specified by ID" versus "specified by
// type+target" distinction in place, nulling whichever attributes do not apply. This keeps
// the planned recipient consistent with what the API will return for it.
//
// A recipient that is still wholly unknown -- likely dependent on the creation of another
// resource -- is left alone.
//
// Shared by the notification recipient plan modifiers so that the Trigger fork and the base
// block cannot drift apart on this logic.
func normalizeRecipientIdentity(r *models.NotificationRecipientModel) {
	if r.ID.IsUnknown() && r.Type.IsUnknown() {
		return
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
}

func NotificationRecipients() planmodifier.Set {
	return notificationRecipientsModifier{}
}
