package honeycombio

import (
	"context"
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/test"
)

func init() {
	resource.AddTestSweepers("recipients", &resource.Sweeper{
		Name: "recipients",
		F: func(_ string) error {
			ctx := context.Background()

			c, err := honeycombio.NewClient()
			if err != nil {
				return fmt.Errorf("could not initialize client: %w", err)
			}
			rcpts, err := c.Recipients.List(ctx)
			if err != nil {
				return fmt.Errorf("could not list recipients: %w", err)
			}

			var name string
			for _, r := range rcpts {
				switch r.Type {
				case honeycombio.RecipientTypeEmail:
					name = r.Details.EmailAddress
				case honeycombio.RecipientTypeSlack:
					name = r.Details.SlackChannel
				case honeycombio.RecipientTypePagerDuty:
					name = r.Details.PDIntegrationName
				case honeycombio.RecipientTypeWebhook, honeycombio.RecipientTypeMSTeams:
					name = r.Details.WebhookName
				default:
					log.Printf("[ERROR] unknown recipient type: %s", r.Type)
					continue
				}

				if strings.HasPrefix(name, "test.") {
					log.Printf("[DEBUG] deleting recipient %s", r.ID)
					err = c.Recipients.Delete(ctx, r.ID)
					if err != nil {
						log.Printf("[ERROR] could not delete recipient %s: %s", r.ID, err)
					}
				}
			}

			return nil
		},
	})
}

func TestAccHoneycombioRecipients_basic(t *testing.T) {
	randomEmail := test.RandomEmail()
	randomSlackChannel := test.RandomStringWithPrefix("#test.", 12)
	randomPDIntegrationKey := test.RandomString(32)
	randomPDIntegrationName := test.RandomStringWithPrefix("test.", 16)
	randomWebhookName := test.RandomStringWithPrefix("test.", 16)
	randomTeamsName := test.RandomStringWithPrefix("test.", 16)

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "honeycombio_email_recipient" "test" {
  address = "%s"
}`, randomEmail),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecipientExists(t, "honeycombio_email_recipient.test"),
					resource.TestCheckResourceAttr("honeycombio_email_recipient.test", "address", randomEmail),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "honeycombio_slack_recipient" "test" {
  channel = "%s"
}`, randomSlackChannel),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecipientExists(t, "honeycombio_slack_recipient.test"),
					resource.TestCheckResourceAttr("honeycombio_slack_recipient.test", "channel", randomSlackChannel),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "honeycombio_pagerduty_recipient" "test" {
  integration_key  = "%s"
  integration_name = "%s"
}`, randomPDIntegrationKey, randomPDIntegrationName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecipientExists(t, "honeycombio_pagerduty_recipient.test"),
					resource.TestCheckResourceAttr("honeycombio_pagerduty_recipient.test", "integration_key", randomPDIntegrationKey),
					resource.TestCheckResourceAttr("honeycombio_pagerduty_recipient.test", "integration_name", randomPDIntegrationName),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "honeycombio_webhook_recipient" "test" {
  name    = "%s"
  secret  = "s0s3kr3t!"
  url     = "https://my.url.corp.net"
}`, randomWebhookName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecipientExists(t, "honeycombio_webhook_recipient.test"),
					resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "name", randomWebhookName),
					resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "secret", "s0s3kr3t!"),
					resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "url", "https://my.url.corp.net"),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "honeycombio_msteams_recipient" "test" {
  name = "%s"
  url  = "https://my.url.office.com/webhooks/incoming/123456789/abcdefg"
}`, randomTeamsName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecipientExists(t, "honeycombio_msteams_recipient.test"),
					resource.TestCheckResourceAttr("honeycombio_msteams_recipient.test", "name", randomTeamsName),
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
