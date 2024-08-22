package modifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

type deletionProtectionModifier struct{}

var _ planmodifier.Bool = &deletionProtectionModifier{}

func (m deletionProtectionModifier) Description(_ context.Context) string {
	return "Ensures delete protection cannot be disabled at creation."
}

func (m deletionProtectionModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m deletionProtectionModifier) PlanModifyBool(ctx context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
	// previous state doesn't exist (creation), and planned value false (disabled)
	if req.State.Raw.IsNull() && !req.PlanValue.ValueBool() {
		resp.Diagnostics.AddAttributeError(
			path.Root("").AtName("delete_protected"),
			"Validation error",
			"Delete protection cannot be disabled at creation.",
		)
	}
}

// EnforceDeletionProtection creates a plan modifier that ensures deletion
// protection cannot be disabled at creation.
func EnforceDeletionProtection() planmodifier.Bool {
	return deletionProtectionModifier{}
}
