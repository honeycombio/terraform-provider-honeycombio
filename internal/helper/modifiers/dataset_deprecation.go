package modifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
)

type datasetDeprecation struct {
	// if true, will allow the previous value of dataset to not strictly be `__all__`
	allowAnyDataset bool
}

var _ planmodifier.String = &datasetDeprecation{}

func (m datasetDeprecation) Description(_ context.Context) string {
	return "Avoids unnecessary replacement if dataset is unset. " +
		"Intended to allow removal of `__all__` from the dataset field, " +
		"but if `allowAnyDataset` is true, will allow any previous value."
}

func (m datasetDeprecation) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m datasetDeprecation) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// Do nothing on resource destroy.
	if req.Plan.Raw.IsNull() {
		return
	}

	if req.ConfigValue.IsNull() {
		// default value is null
		resp.PlanValue = types.StringNull()
	}

	// null now, but what was it previously?
	if req.ConfigValue.IsNull() {
		if m.allowAnyDataset || req.StateValue.ValueString() == client.EnvironmentWideSlug {
			// if the previous value was `__all__`, or we're allowing any previous value, suppress the diff
			resp.PlanValue = req.StateValue
			return
		}
	}

	if !req.PlanValue.Equal(req.StateValue) {
		// require replacement only if the dataset value is otherwise changing
		resp.RequiresReplace = true
		return
	}

}

// DatasetDeprecation avoids unnecessary diffs if dataset becomes omitted.
// Intended to allow removal of `__all__` from the dataset field,
// but if `allowAnyDataset` is true, will allow any previous value of dataset.
func DatasetDeprecation(allowAnyDataset bool) planmodifier.String {
	return datasetDeprecation{
		allowAnyDataset: allowAnyDataset,
	}
}
