package honeycombio

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	honeycombio "github.com/kvrhdn/go-honeycombio"
)

func TestAccDataSourceHoneycombioTriggerRecipient_basic(t *testing.T) {
	c := testAccProvider.Meta().(*honeycombio.Client)
	dataset := testAccDataset()

	trigger := testAccTriggerRecipientCreateTrigger(t, c, dataset)
	defer testAccTriggerRecipientDeleteTrigger(t, c, dataset, trigger)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTriggerRecipient(dataset, "email", "acctest@example.com"),
			},
			{
				Config:      testAccTriggerRecipient(dataset, "email", "another@example.com"),
				ExpectError: regexp.MustCompile("could not find a trigger recipient in "),
			},
			{
				Config:      testAccTriggerRecipient(dataset, "slack", "acctest@example.com"),
				ExpectError: regexp.MustCompile("could not find a trigger recipient in "),
			},
		},
	})
}

func testAccTriggerRecipient(dataset, recipientType, target string) string {
	return fmt.Sprintf(`
data "honeycombio_trigger_recipient" "test" {
  dataset = "%s"
  type = "%s"
  target = "%s"
}`, dataset, recipientType, target)
}

func testAccTriggerRecipientCreateTrigger(t *testing.T, c *honeycombio.Client, dataset string) *honeycombio.Trigger {
	trigger := &honeycombio.Trigger{
		Name: "Terraform provider - acc test trigger recipient",
		Query: &honeycombio.QuerySpec{
			Calculations: []honeycombio.CalculationSpec{
				{
					Op: honeycombio.CalculateOpCount,
				},
			},
		},
		Threshold: &honeycombio.TriggerThreshold{
			Op:    honeycombio.TriggerThresholdOpGreaterThan,
			Value: &[]float64{100}[0],
		},
		Recipients: []honeycombio.TriggerRecipient{
			{
				Type:   honeycombio.TriggerRecipientTypeEmail,
				Target: "acctest@example.com",
			},
		},
	}
	trigger, err := c.Triggers.Create(context.Background(), dataset, trigger)
	if err != nil {
		t.Error(err)
	}
	return trigger
}

func testAccTriggerRecipientDeleteTrigger(t *testing.T, c *honeycombio.Client, dataset string, trigger *honeycombio.Trigger) {
	err := c.Triggers.Delete(context.Background(), dataset, trigger.ID)
	if err != nil {
		t.Error(err)
	}
}
