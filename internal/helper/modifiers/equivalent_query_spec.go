package modifiers

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
)

type equivalentQuerySpec struct{}

var _ planmodifier.String = &equivalentQuerySpec{}

func (m equivalentQuerySpec) Description(_ context.Context) string {
	return "Avoids unnecessary plans if two query specifications are equivalent."
}

func (m equivalentQuerySpec) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m equivalentQuerySpec) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// Do nothing on resource destroy.
	if req.Plan.Raw.IsNull() {
		return
	}
	// Do nothing if the plan or state is not yet known.
	if req.PlanValue.IsUnknown() || req.PlanValue.IsNull() || req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		return
	}

	var configQs, stateQs client.QuerySpec
	config := req.ConfigValue.ValueString()
	if err := json.Unmarshal([]byte(config), &configQs); err != nil {
		resp.Diagnostics.AddError("Error unmarshaling config JSON", err.Error())
		return
	}
	state := req.StateValue.ValueString()
	if err := json.Unmarshal([]byte(state), &stateQs); err != nil {
		resp.Diagnostics.AddError("Error unmarshaling state JSON", err.Error())
		return
	}

	if stateQs.EquivalentTo(configQs) {
		// If the query specifications are equivalent, suppress the diff
		// by setting the plan value to value already in state.
		resp.PlanValue = req.StateValue
	}
}

// EquivalentQuerySpec avoids unnecessary plans if two query specifications are equivalent.
func EquivalentQuerySpec() planmodifier.String {
	return equivalentQuerySpec{}
}
