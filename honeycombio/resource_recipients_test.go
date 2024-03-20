package honeycombio

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccHoneycombioRecipients_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
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
			{
				Config: `
resource "honeycombio_slack_recipient" "test" {
  channel = "#alerty-alerts"
}
`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecipientExists(t, "honeycombio_slack_recipient.test"),
					resource.TestCheckResourceAttr("honeycombio_slack_recipient.test", "channel", "#alerty-alerts"),
				),
			},
			{
				Config: `
resource "honeycombio_pagerduty_recipient" "test" {
  integration_key  = "cd6e8de3c857aefc950e0d5ebcb79ac2"
  integration_name = "myservice-notifications"
}
`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecipientExists(t, "honeycombio_pagerduty_recipient.test"),
					resource.TestCheckResourceAttr("honeycombio_pagerduty_recipient.test", "integration_key", "cd6e8de3c857aefc950e0d5ebcb79ac2"),
					resource.TestCheckResourceAttr("honeycombio_pagerduty_recipient.test", "integration_name", "myservice-notifications"),
				),
			},
			{
				Config: `
resource "honeycombio_webhook_recipient" "test" {
  name    = "custom-alert-router"
  secret  = "s0s3kr3t!"
  url     = "https://my.url.corp.net"
}
`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecipientExists(t, "honeycombio_webhook_recipient.test"),
					resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "name", "custom-alert-router"),
					resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "secret", "s0s3kr3t!"),
					resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "url", "https://my.url.corp.net"),
				),
			},
			{
				Config: `
			resource "honeycombio_msteams_recipient" "test" {
			  name = "my teams channel"
			  url  = "https://my.url.office.com/webhooks/incoming/123456789/abcdefg"
			}
			`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecipientExists(t, "honeycombio_msteams_recipient.test"),
					resource.TestCheckResourceAttr("honeycombio_msteams_recipient.test", "name", "my teams channel"),
					resource.TestCheckResourceAttr("honeycombio_msteams_recipient.test", "url", "https://my.url.office.com/webhooks/incoming/123456789/abcdefg"),
				),
			},
		},
	})
}

func testAccCheckRecipientExists(t *testing.T, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resourceState, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		client := testAccClient(t)
		_, err := client.Recipients.Get(context.Background(), resourceState.Primary.ID)
		if err != nil {
			return fmt.Errorf("could not find created Recipient: %w", err)
		}

		return nil
	}
}
