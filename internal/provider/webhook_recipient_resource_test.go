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
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "variable.#", "0"),
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
		createBody := `<<EOT
		{
			"name": " {{ .Name }}",
			"id": " {{ .ID }}",
			"description": " {{ .Description }}",
		}
		EOT`
		updateBody := `<<EOT
		{
			"name": " {{ .Name }}",
			"id": " {{ .ID }}"
		}
		EOT`

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

	variable {
	  name = "severity"
      default_value = "critical"
	}

	template {
	  type   = "trigger"
      body = %s
    }
}`, name, url, createBody),
					Check: resource.ComposeAggregateTestCheckFunc(
						testAccEnsureRecipientExists(t, "honeycombio_webhook_recipient.test"),
						resource.TestCheckResourceAttrSet("honeycombio_webhook_recipient.test", "id"),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "name", name),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "url", url),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "template.#", "1"),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "template.0.type", "trigger"),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "variable.#", "1"),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "variable.0.name", "severity"),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "variable.0.default_value", "critical"),
						resource.TestCheckNoResourceAttr("honeycombio_webhook_recipient.test", "secret"),
					),
				},
				{
					Config: fmt.Sprintf(`
resource "honeycombio_webhook_recipient" "test" {
  name = "%s"
	url  = "%s"

	secret = "so-secret"

	variable {
	  name = "severity"
      default_value = "warning"
	}

	template {
	  type   = "trigger"
      body = %s
    }
}`, name, url, updateBody),
					Check: resource.ComposeAggregateTestCheckFunc(
						testAccEnsureRecipientExists(t, "honeycombio_webhook_recipient.test"),
						resource.TestCheckResourceAttrSet("honeycombio_webhook_recipient.test", "id"),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "name", name),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "url", url),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "secret", "so-secret"),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "template.#", "1"),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "template.0.type", "trigger"),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "variable.#", "1"),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "variable.0.name", "severity"),
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
		createBody := `<<EOT
		{
			"name": " {{ .Name }}",
			"id": " {{ .ID }}",
			"description": " {{ .Description }}",
		}
		EOT`
		updateBody := `<<EOT
		{
			"name": " {{ .Name }}",
			"id": " {{ .ID }}"
		}
		EOT`

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

	variable {
	  name = "severity"
      default_value = "critical"
	}

	template {
	  type   = "trigger"
      body = %s
    }
}`, name, url, createBody),
					Check: resource.ComposeAggregateTestCheckFunc(
						testAccEnsureRecipientExists(t, "honeycombio_webhook_recipient.test"),
						resource.TestCheckResourceAttrSet("honeycombio_webhook_recipient.test", "id"),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "name", name),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "url", url),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "template.#", "1"),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "template.0.type", "trigger"),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "variable.#", "1"),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "variable.0.name", "severity"),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "variable.0.default_value", "critical"),
						resource.TestCheckNoResourceAttr("honeycombio_webhook_recipient.test", "secret"),
					),
				},
				{
					Config: fmt.Sprintf(`
resource "honeycombio_webhook_recipient" "test" {
  name = "%s"
	url  = "%s"

	secret = "so-secret"

	variable {
	  name = "severity"
      default_value = "warning"
	}

	template {
	  type   = "trigger"
      body = %s
    }
}`, name, url, updateBody),
					Check: resource.ComposeAggregateTestCheckFunc(
						testAccEnsureRecipientExists(t, "honeycombio_webhook_recipient.test"),
						resource.TestCheckResourceAttrSet("honeycombio_webhook_recipient.test", "id"),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "name", name),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "url", url),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "secret", "so-secret"),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "template.#", "1"),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "template.0.type", "trigger"),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "variable.#", "1"),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "variable.0.name", "severity"),
					),
				},
				{
					ResourceName: "honeycombio_webhook_recipient.test",
					ImportState:  true,
				},
			},
		})
	})

	t.Run("custom webhook succeeds when a template is removed", func(t *testing.T) {
		name := test.RandomStringWithPrefix("test.", 20)
		url := test.RandomURL()
		body := `<<EOT
		{
			"name": " {{ .Name }}",
			"id": " {{ .ID }}"
		}
		EOT`

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
      body = %s
    }
}`, name, url, body),
					Check: resource.ComposeAggregateTestCheckFunc(
						testAccEnsureRecipientExists(t, "honeycombio_webhook_recipient.test"),
						resource.TestCheckResourceAttrSet("honeycombio_webhook_recipient.test", "id"),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "name", name),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "url", url),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "template.#", "1"),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "template.0.type", "trigger"),
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

	t.Run("custom webhook succeeds when a variable is removed", func(t *testing.T) {
		name := test.RandomStringWithPrefix("test.", 20)
		url := test.RandomURL()
		body := `<<EOT
		{
			"name": " {{ .Name }}",
			"id": " {{ .ID }}"
		}
		EOT`

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

	variable {
	  name = "severity"
      default_value = "critical"
	}

	template {
	  type   = "trigger"
      body = %s
    }
}`, name, url, body),
					Check: resource.ComposeAggregateTestCheckFunc(
						testAccEnsureRecipientExists(t, "honeycombio_webhook_recipient.test"),
						resource.TestCheckResourceAttrSet("honeycombio_webhook_recipient.test", "id"),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "name", name),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "url", url),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "template.#", "1"),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "template.0.type", "trigger"),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "variable.#", "1"),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "variable.0.name", "severity"),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "variable.0.default_value", "critical"),
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
      body = %s
    }
}`, name, url, body),
					Check: resource.ComposeAggregateTestCheckFunc(
						testAccEnsureRecipientExists(t, "honeycombio_webhook_recipient.test"),
						resource.TestCheckResourceAttrSet("honeycombio_webhook_recipient.test", "id"),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "name", name),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "url", url),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "secret", "so-secret"),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "template.#", "1"),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "template.0.type", "trigger"),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "variable.#", "0"),
					),
				},
				{
					ResourceName: "honeycombio_webhook_recipient.test",
					ImportState:  true,
				},
			},
		})
	})

	t.Run("custom webhook succeeds when a variable has no default value", func(t *testing.T) {
		name := test.RandomStringWithPrefix("test.", 20)
		url := test.RandomURL()
		body := `<<EOT
		{
			"name": " {{ .Name }}",
			"id": " {{ .ID }}"
		}
		EOT`

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

	variable {
	  name = "variable1"
      default_value = ""
	}

	variable {
	  name = "variable2"
      default_value = "critical"
	}

	template {
	  type   = "trigger"
      body = %s
    }
}`, name, url, body),
					Check: resource.ComposeAggregateTestCheckFunc(
						testAccEnsureRecipientExists(t, "honeycombio_webhook_recipient.test"),
						resource.TestCheckResourceAttrSet("honeycombio_webhook_recipient.test", "id"),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "name", name),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "url", url),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "template.#", "1"),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "template.0.type", "trigger"),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "variable.#", "2"),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "variable.0.name", "variable1"),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "variable.0.default_value", ""),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "variable.1.name", "variable2"),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "variable.1.default_value", "critical"),
						resource.TestCheckNoResourceAttr("honeycombio_webhook_recipient.test", "secret"),
					),
				},
				{
					Config: fmt.Sprintf(`
resource "honeycombio_webhook_recipient" "test" {
  name = "%s"
	url  = "%s"

	variable {
	  name = "variable1"
      default_value = ""
	}

	variable {
	  name = "variable2"
      default_value = ""
	}

	template {
	  type   = "trigger"
      body = %s
    }
}`, name, url, body),
					Check: resource.ComposeAggregateTestCheckFunc(
						testAccEnsureRecipientExists(t, "honeycombio_webhook_recipient.test"),
						resource.TestCheckResourceAttrSet("honeycombio_webhook_recipient.test", "id"),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "name", name),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "url", url),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "template.#", "1"),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "template.0.type", "trigger"),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "variable.#", "2"),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "variable.0.name", "variable1"),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "variable.0.default_value", ""),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "variable.1.name", "variable2"),
						resource.TestCheckResourceAttr("honeycombio_webhook_recipient.test", "variable.1.default_value", ""),
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
					ExpectError: regexp.MustCompile(`cannot have more than one "template" of type "trigger"`),
				},
			},
		})

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

	variable {
	  name   = "severity"
      default_value = "critical"
    }

	variable {
	  name   = "severity"
      default_value = "warning"
    }
}`, name, url),
					ExpectError: regexp.MustCompile(`cannot have more than one "variable" with the same "name"`),
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
