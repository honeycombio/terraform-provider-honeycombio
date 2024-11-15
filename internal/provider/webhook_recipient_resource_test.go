package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/test"
)

func TestAcc_WebhookRecipientResource(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		name := test.RandomStringWithPrefix("test.", 20)
		url := test.RandomURL()

		resource.Test(t, resource.TestCase{
			PreCheck:                 testAccPreCheck(t),
			ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
			Steps: []resource.TestStep{
				{
					Config: fmt.Sprintf(`
resource "honeycombio_webhook_recipient" "test" {
  name = "%s"
	url  = "%s"
}`, name, url),
					Check: resource.ComposeAggregateTestCheckFunc(
						testAccEnsureRecipientExists(t, "honeycombio_webhook_recipient.test"),
						resource.TestCheckResourceAttrSet("honeycombio_webhook_recipient.test", "id"),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "name", name),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "url", url),
						resource.TestCheckNoResourceAttr("honeycombio_webhook_recipient.test", "secret"),
					),
				},
				{
					Config: fmt.Sprintf(`
resource "honeycombio_webhook_recipient" "test" {
  name = "%s"
	url  = "%s"

	secret = "so-secret"
}`, name, url),
					Check: resource.ComposeAggregateTestCheckFunc(
						testAccEnsureRecipientExists(t, "honeycombio_webhook_recipient.test"),
						resource.TestCheckResourceAttrSet("honeycombio_webhook_recipient.test", "id"),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "name", name),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "url", url),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "secret", "so-secret"),
					),
				},
				{
					ResourceName: "honeycombio_webhook_recipient.test",
					ImportState:  true,
				},
			},
		})
	})
}

// TestAcc_WebhookRecipientResource_UpgradeFromVersion027 tests the migration case from the
// last SDK-based version of the Webhook Recipient resource to the current Framework-based version.
//
// See: https://developer.hashicorp.com/terraform/plugin/framework/migrating/testing#testing-migration
func TestAcc_WebhookRecipientResource_UpgradeFromVersion027(t *testing.T) {
	name := test.RandomStringWithPrefix("test.", 20)
	url := test.RandomURL()
	config := fmt.Sprintf(`
resource "honeycombio_webhook_recipient" "test" {
  name   = "%s"
	url    = "%s"
	secret = "so-secret"
}`, name, url)

	resource.Test(t, resource.TestCase{
		PreCheck: testAccPreCheck(t),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"honeycombio": {
						VersionConstraint: "0.27",
						Source:            "honeycombio/honeycombio",
					},
				},
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccEnsureRecipientExists(t, "honeycombio_webhook_recipient.test"),
				),
			},
			{
				ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
				Config:                   config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("honeycombio_webhook_recipient.test", "id"),
					resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "name", name),
					resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "url", url),
					resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "secret", "so-secret"),
				),
			},
		},
	})
}

func testAccEnsureRecipientExists(t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("resource not found: %s", name)
		}

		client := testAccClient(t)
		_, err := client.Recipients.Get(context.Background(), rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("failed to fetch created recipient: %s", err)
		}

		return nil
	}
}
