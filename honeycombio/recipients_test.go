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
					ExpectError: regexp.MustCompile(`Creating new MSTeams recipients is no longer possible`),
				},
			},
		})
	})
}
