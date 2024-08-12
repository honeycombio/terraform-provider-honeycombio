package honeycombio

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/test"
)

// TestAccHoneycombMSTeamsRecipient tests the creation
// and validation of the original Honeycomb MSTeams recipient
// and the new Honeycomb MSTeams Workflow recipient.
func TestAccHoneycombMSTeamsRecipient(t *testing.T) {
	t.Run("workflow recipient works", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:                 testAccPreCheck(t),
			ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
			Steps: []resource.TestStep{
				{
					Config: fmt.Sprintf(`
resource "honeycombio_msteams_workflow_recipient" "test" {
  name = "%s"
  url  = "https://example.com"
}`, test.RandomStringWithPrefix("test.", 10)),
				},
			},
		})
	})

	t.Run("new webhook recipient fails creation", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:                 testAccPreCheck(t),
			ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
			Steps: []resource.TestStep{
				{
					Config: `
resource "honeycombio_msteams_recipient" "test" {
  name = "test"
  url  = "https://example.com"
}`,
					ExpectError: regexp.MustCompile(`recipient creation is no longer allowed`),
				},
			},
		})
	})

	t.Run("webhook recipient created with earlier version can be further managed", func(t *testing.T) {
		config := func(name string) string {
			return fmt.Sprintf(`
resource "honeycombio_msteams_recipient" "test" {
  name = "%s"
  url  = "https://mycorp.example.net/webhooks/incoming/12345"
}`, name)
		}

		resource.Test(t, resource.TestCase{
			Steps: []resource.TestStep{
				{
					// create the recipient with v0.25.0 of the provider
					ExternalProviders: map[string]resource.ExternalProvider{
						"honeycombio": {
							VersionConstraint: "0.25.0", // last version before the deprecation
							Source:            "honeycombio/honeycombio",
						},
					},
					Config: config(test.RandomStringWithPrefix("test.", 10)),
				},
				// update the recipient's name with the latest version of the provider
				// and then allow the test to clean up the resource
				{
					ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
					Config:                   config(test.RandomStringWithPrefix("test.", 10)),
				},
			},
		})
	})
}
