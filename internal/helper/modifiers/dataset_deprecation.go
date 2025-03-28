package modifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type datasetDeprecation struct{}

var _ planmodifier.String = &datasetDeprecation{}

func (m datasetDeprecation) Description(_ context.Context) string {
	return "Avoids unnecessary plans if dataset becomes omitted. Configuration should now behave the same."
}

func (m datasetDeprecation) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m datasetDeprecation) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// Do nothing on resource destroy.
	if req.Plan.Raw.IsNull() {
		return
	}
	// Assign null value if the plan value is unknown or theres no state value to fall back on
	if req.PlanValue.IsUnknown() || req.StateValue.IsUnknown() {
		resp.PlanValue = types.StringNull()
		return
	}

	// Suppress diff if the old and new values are equal
	if req.PlanValue.Equal(req.StateValue) {
		resp.PlanValue = req.StateValue
		return
	}
	// Suppress diff if the new value is empty
	if req.PlanValue.IsNull() {
		resp.PlanValue = req.StateValue
		return
	}
}

// DatasetDeprecation avoids unnecessary plans if dataset becomes omitted. Configuration should now behave the same.
func DatasetDeprecation() planmodifier.String {
	return datasetDeprecation{}
}
