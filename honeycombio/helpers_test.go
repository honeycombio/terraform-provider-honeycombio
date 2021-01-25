package honeycombio

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/kvrhdn/go-honeycombio"
)

// testCheckOutputContains checks an output in the Terraform configuration
// contains the given value. The output is expected to be of type list(string).
func testCheckOutputContains(name, contains string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ms := s.RootModule()
		rs, ok := ms.Outputs[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		output := rs.Value.([]interface{})

		for _, value := range output {
			if value.(string) == contains {
				return nil
			}
		}

		return fmt.Errorf("Output '%s' did not contain %#v, got %#v", name, contains, output)
	}
}

// testCheckOutputDoesNotContain checks an output in the Terraform configuration
// does not contain the given value. The output is expected to be of type
// list(string).
func testCheckOutputDoesNotContain(name, contains string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ms := s.RootModule()
		rs, ok := ms.Outputs[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		output := rs.Value.([]interface{})

		for _, value := range output {
			if value.(string) == contains {
				return fmt.Errorf("Output '%s' contained %#v, should not", name, contains)
			}
		}

		return nil
	}
}

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
			Value: 100,
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
