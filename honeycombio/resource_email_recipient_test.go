package honeycombio

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccHoneycombioEmailRecipient_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          testAccPreCheck(t),
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "honeycombio_email_recipient" "test" {
  address = "tf-test@example.com"
}
`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecipientExists(t, "honeycombio_email_recipient.test"),
					resource.TestCheckResourceAttr("honeycombio_email_recipient.test", "address", "tf-test@example.com"),
				),
			},
		},
	})
}
