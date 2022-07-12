package honeycombio

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
)

func TestAccDataSourceHoneycombioTriggerRecipient_basic(t *testing.T) {
	dataset := testAccDataset()

	_, deleteFn := createTriggerWithRecipient(t, dataset, honeycombio.NotificationRecipient{
		Type:   honeycombio.RecipientTypeEmail,
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
				ExpectError: regexp.MustCompile("could not find a trigger recipient in .* with type = \"email\" and target = \"another@example.com\""),
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
  type = "%s"
  target = "%s"
}`, dataset, recipientType, target)
}
