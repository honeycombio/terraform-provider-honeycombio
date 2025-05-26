package modifiers

import (
	"context"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

type equivalentTags struct{}

var _ planmodifier.Map = &equivalentTags{}

func (m equivalentTags) Description(_ context.Context) string {
	return "Suppresses unnecessary diffs when tag maps are equivalent."
}

func (m equivalentTags) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m equivalentTags) PlanModifyMap(ctx context.Context, req planmodifier.MapRequest, resp *planmodifier.MapResponse) {
	// Do nothing if plan or state is not yet known or is null
	if req.PlanValue.IsUnknown() || req.StateValue.IsUnknown() || req.PlanValue.IsNull() || req.StateValue.IsNull() {
		return
	}

	var stateMap, planMap map[string]string
	diags := req.StateValue.ElementsAs(ctx, &stateMap, false)
	if diags.HasError() {
		return
	}

	diags = req.PlanValue.ElementsAs(ctx, &planMap, false)
	if diags.HasError() {
		return
	}

	// If the maps are equivalent, suppress the diff
	if reflect.DeepEqual(stateMap, planMap) {
		resp.PlanValue = req.StateValue
		return
	}
}

// EquivalentTags avoids unnecessary diffs when tag maps are equivalent.
func EquivalentTags() planmodifier.Map {
	return equivalentTags{}
}
