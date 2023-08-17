package modifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type defaultTriggerThresholdExceededValueModifier struct{}

var _ planmodifier.Int64 = &defaultTriggerThresholdExceededValueModifier{}

func (m defaultTriggerThresholdExceededValueModifier) Description(_ context.Context) string {
	return "Handles the default value for a Trigger's Threshold Exceeded Limits."
}

func (m defaultTriggerThresholdExceededValueModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m defaultTriggerThresholdExceededValueModifier) PlanModifyInt64(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	// Do nothing on resource destroy.
	if req.Plan.Raw.IsNull() {
		return
	}

	if req.ConfigValue.IsNull() {
		// default value is 1
		resp.PlanValue = types.Int64Value(1)
	}
}

func DefaultTriggerThresholdExceededValue() planmodifier.Int64 {
	return defaultTriggerThresholdExceededValueModifier{}
}
