package honeycombio

import (
	"context"
	"testing"

	"github.com/kvrhdn/go-honeycombio"
)

func createTriggerWithRecipient(t *testing.T, dataset string, recipient honeycombio.TriggerRecipient) (trigger *honeycombio.Trigger, deleteFn func()) {
	ctx := context.Background()
	c := testAccClient(t)

	trigger = &honeycombio.Trigger{
		Name: "Terraform provider - acc test trigger recipient",
		Query: &honeycombio.QuerySpec{
			Calculations: []honeycombio.CalculationSpec{
				{
					Op: honeycombio.CalculationOpCount,
				},
			},
		},
		Threshold: &honeycombio.TriggerThreshold{
			Op:    honeycombio.TriggerThresholdOpGreaterThan,
			Value: honeycombio.Float64Ptr(100),
		},
		Recipients: []honeycombio.TriggerRecipient{recipient},
	}
	trigger, err := c.Triggers.Create(ctx, dataset, trigger)
	if err != nil {
		t.Error(err)
	}

	return trigger, func() {
		err := c.Triggers.Delete(ctx, dataset, trigger.ID)
		if err != nil {
			t.Error(err)
		}
	}
}
