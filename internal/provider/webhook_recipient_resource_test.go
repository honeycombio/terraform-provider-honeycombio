package provider

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/test"
)

func TestAcc_WebhookRecipientResource(t *testing.T) {
	t.Run("happy path standard webhook", func(t *testing.T) {
		name := test.RandomStringWithPrefix("test.", 20)
		url := test.RandomURL()

		resource.Test(t, resource.TestCase{
			PreCheck:                 testAccPreCheck(t),
			ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
			CheckDestroy:             testAccEnsureRecipientDestroyed(t),
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
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "template.#", "0"),
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
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "template.#", "0"),
					),
				},
				{
					ResourceName: "honeycombio_webhook_recipient.test",
					ImportState:  true,
				},
			},
		})
	})

	t.Run("happy path custom webhook", func(t *testing.T) {
		name := test.RandomStringWithPrefix("test.", 20)
		url := test.RandomURL()

		resource.Test(t, resource.TestCase{
			PreCheck:                 testAccPreCheck(t),
			ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
			CheckDestroy:             testAccEnsureRecipientDestroyed(t),
			Steps: []resource.TestStep{
				{
					Config: fmt.Sprintf(`
resource "honeycombio_webhook_recipient" "test" {
  name = "%s"
	url  = "%s"

	template {
	  type   = "trigger"
      body = "body"
    }
}`, name, url),
					Check: resource.ComposeAggregateTestCheckFunc(
						testAccEnsureRecipientExists(t, "honeycombio_webhook_recipient.test"),
						resource.TestCheckResourceAttrSet("honeycombio_webhook_recipient.test", "id"),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "name", name),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "url", url),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "template.#", "1"),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "template.0.type", "trigger"),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "template.0.body", "body"),
						resource.TestCheckNoResourceAttr("honeycombio_webhook_recipient.test", "secret"),
					),
				},
				{
					Config: fmt.Sprintf(`
resource "honeycombio_webhook_recipient" "test" {
  name = "%s"
	url  = "%s"

	secret = "so-secret"

	template {
	  type   = "trigger"
      body = "body"
    }
}`, name, url),
					Check: resource.ComposeAggregateTestCheckFunc(
						testAccEnsureRecipientExists(t, "honeycombio_webhook_recipient.test"),
						resource.TestCheckResourceAttrSet("honeycombio_webhook_recipient.test", "id"),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "name", name),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "url", url),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "secret", "so-secret"),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "template.#", "1"),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "template.0.type", "trigger"),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "template.0.body", "body"),
					),
				},
				{
					ResourceName: "honeycombio_webhook_recipient.test",
					ImportState:  true,
				},
			},
		})
	})

	t.Run("custom webhook validations error when they should", func(t *testing.T) {
		name := test.RandomStringWithPrefix("test.", 20)
		url := test.RandomURL()

		resource.Test(t, resource.TestCase{
			PreCheck:                 testAccPreCheck(t),
			ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
			CheckDestroy:             testAccEnsureRecipientDestroyed(t),
			Steps: []resource.TestStep{
				{
					Config: fmt.Sprintf(`
resource "honeycombio_webhook_recipient" "test" {
  name = "%s"
	url  = "%s"

	template {
	  type   = "trigger"
      body = "body"
    }

	template {
	  type   = "trigger"
      body = "another body"
    }
}`, name, url),
					ExpectError: regexp.MustCompile("Conflicting configuration arguments"),
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

func testAccEnsureRecipientDestroyed(t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, resourceState := range s.RootModule().Resources {
			if resourceState.Type != "honeycombio_webhook_recipient" {
				continue
			}

			if resourceState.Primary.ID == "" {
				return fmt.Errorf("no ID set for recipient")
			}

			client := testAccClient(t)
			_, err := client.Recipients.Get(context.Background(), resourceState.Primary.ID)
			if err == nil {
				return fmt.Errorf("recipient %s was not deleted on destroy", resourceState.Primary.ID)
			}
		}

		return nil
	}
}
