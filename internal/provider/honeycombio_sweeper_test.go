package provider

import (
	"context"
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
)

// TestMain is responsible for parsing the special test flags and invoking the sweepers
//
//	See: https://developer.hashicorp.com/terraform/plugin/testing/acceptance-tests/sweepers
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func init() {
	resource.AddTestSweepers("recipients", getRecipientSweeper("recipients"))
}

// getRecipientSweeper returns a Sweeper that deletes recipients with names
// starting with "test." or "#test."
func getRecipientSweeper(name string) *resource.Sweeper {
	return &resource.Sweeper{
		Name: name,
		F: func(_ string) error {
			ctx := context.Background()
			c, err := client.NewClient()
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
				case client.RecipientTypeEmail:
					name = r.Details.EmailAddress
				case client.RecipientTypeSlack:
					name = r.Details.SlackChannel
				case client.RecipientTypePagerDuty:
					name = r.Details.PDIntegrationName
				case client.RecipientTypeWebhook, client.RecipientTypeMSTeams:
					name = r.Details.WebhookName
				default:
					log.Printf("[ERROR] unknown recipient type: %s", r.Type)
					continue
				}

				if strings.HasPrefix(name, "test.") || strings.HasPrefix(name, "#test.") {
					log.Printf("[DEBUG] deleting %s recipient \"%s\" (%s)", r.Type, name, r.ID)
					err = c.Recipients.Delete(ctx, r.ID)
					if err != nil {
						log.Printf("[ERROR] could not delete %s recipient %s: %s", r.Type, r.ID, err)
					}
				}
			}

			return nil
		},
	}
}
