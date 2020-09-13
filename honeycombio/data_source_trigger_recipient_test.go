package honeycombio

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/kvrhdn/go-honeycombio"
)

func TestAccDataSourceHoneycombioTriggerRecipient_basic(t *testing.T) {
	c := testAccClient(t)
	dataset := testAccDataset()

	_, deleteFn := createTriggerWithRecipient(t, c, dataset, honeycombio.TriggerRecipient{
		Type:   honeycombio.TriggerRecipientTypeEmail,
		Target: "acctest@example.com",
	})
	defer deleteFn()

	resource.Test(t, resource.TestCase{
		PreCheck:          testAccPreCheck(t),
		ProviderFactories: testAccProviderFactories,
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
