package honeycombio

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/test"
)

func TestAccDataSourceHoneycombioTriggerRecipient_basic(t *testing.T) {
	ctx := context.Background()
	dataset := testAccDataset()
	c := testAccClient(t)

	randomEmail := test.RandomEmail()
	trigger, err := c.Triggers.Create(ctx, dataset, &honeycombio.Trigger{
		Name: test.RandomStringWithPrefix("test.", 12),
		Query: &honeycombio.QuerySpec{
			Calculations: []honeycombio.CalculationSpec{
				{Op: honeycombio.CalculationOpCount},
			},
		},
		Threshold: &honeycombio.TriggerThreshold{
			Op:    honeycombio.TriggerThresholdOpGreaterThan,
			Value: 100,
		},
		Recipients: []honeycombio.NotificationRecipient{
			{
				Type:   honeycombio.RecipientTypeEmail,
				Target: randomEmail,
			},
		},
	})
	require.NoError(t, err)
	//nolint:errcheck
	t.Cleanup(func() {
		c.Triggers.Delete(ctx, dataset, trigger.ID)
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
		Steps: []resource.TestStep{
			{
				Config: testAccTriggerRecipient(dataset, "email", randomEmail),
			},
			{
				Config:      testAccTriggerRecipient(dataset, "email", "random@corp.net"),
				ExpectError: regexp.MustCompile("could not find a trigger recipient in .* with type = \"email\" and target = \"random@corp.net\""),
			},
			{
				Config:      testAccTriggerRecipient(dataset, "slack", "honeycombio"),
				ExpectError: regexp.MustCompile("could not find a trigger recipient in .* with type = \"slack\" and target = \"honeycombio\""),
			},
		},
	})
}

func testAccTriggerRecipient(dataset, recipientType, target string) string {
	return fmt.Sprintf(`
data "honeycombio_trigger_recipient" "test" {
  dataset = "%s"
  type    = "%s"
  target  = "%s"
}`, dataset, recipientType, target)
}
