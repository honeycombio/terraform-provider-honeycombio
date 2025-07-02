package modifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// RequiresReplaceIfDatasetChanges returns a plan modifier that forces resource replacement
// when the datasets field changes.
func RequiresReplaceIfDatasetChanges() planmodifier.Set {
	return requiresReplaceIfDatasetChangesSetModifier{}
}

var _ planmodifier.Set = &requiresReplaceIfDatasetChangesSetModifier{}

type requiresReplaceIfDatasetChangesSetModifier struct{}

func (r requiresReplaceIfDatasetChangesSetModifier) Description(ctx context.Context) string {
	return "If the datasets value changes, Terraform will destroy and recreate the resource."
}

func (r requiresReplaceIfDatasetChangesSetModifier) MarkdownDescription(ctx context.Context) string {
	return r.Description(ctx)
}

func (r requiresReplaceIfDatasetChangesSetModifier) PlanModifySet(ctx context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
	// Don't require replacement during create
	if req.State.Raw.IsNull() {
		return
	}

	// Don't require replacement during destroy
	if req.Plan.Raw.IsNull() {
		return
	}

	var stateValue, planValue types.Set

	resp.Diagnostics.Append(req.State.GetAttribute(ctx, req.Path, &stateValue)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, req.Path, &planValue)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If both are null, no change
	if stateValue.IsNull() && planValue.IsNull() {
		return
	}

	// If both are unknown, can't determine change
	if stateValue.IsUnknown() || planValue.IsUnknown() {
		return
	}

	// If values are equal, no change
	if stateValue.Equal(planValue) {
		return
	}

	// If we get here, the datasets value has changed - require replacement
	resp.RequiresReplace = true
}
