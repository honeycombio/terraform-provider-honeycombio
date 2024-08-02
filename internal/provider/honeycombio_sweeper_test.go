package provider

import (
	"context"
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
	v2client "github.com/honeycombio/terraform-provider-honeycombio/client/v2"
)

const (
	// SweeperTargetPrefix is the prefix used to identify resources which
	// will be deleted by the sweepers
	SweeperTargetPrefix = "test."
)

// TestMain is responsible for parsing the special test flags and invoking the sweepers
//
//	See: https://developer.hashicorp.com/terraform/plugin/testing/acceptance-tests/sweepers
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func init() {
	resource.AddTestSweepers("environments", getEnvironmentSweeper("environments"))
	resource.AddTestSweepers("recipients", getRecipientSweeper("recipients"))
}

func getEnvironmentSweeper(name string) *resource.Sweeper {
	return &resource.Sweeper{
		Name: name,
		F: func(_ string) error {
			ctx := context.Background()
			c, err := v2client.NewClient()
			if err != nil {
				return fmt.Errorf("could not initialize client: %w", err)
			}
			pager, err := c.Environments.List(ctx)
			if err != nil {
				return fmt.Errorf("could not list environments: %w", err)
			}

			envs := make([]*v2client.Environment, 0)
			for pager.HasNext() {
				items, err := pager.Next(ctx)
				if err != nil {
					return fmt.Errorf("error listing environments: %w", err)
				}
				envs = append(envs, items...)
			}

			for _, e := range envs {
				if strings.HasPrefix(e.Name, SweeperTargetPrefix) {
					log.Printf("[DEBUG] deleting environment %s (%s)", e.Name, e.ID)
					c.Environments.Update(ctx, &v2client.Environment{
						ID: e.ID,
						Settings: &v2client.EnvironmentSettings{
							DeleteProtected: client.ToPtr(false),
						},
					})
					err = c.Environments.Delete(ctx, e.ID)
					if err != nil {
						log.Printf("[ERROR] could not delete environment %s: %s", e.ID, err)
					}
				}
			}

			return nil
		},
	}
}

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

				if strings.HasPrefix(name, "#"+SweeperTargetPrefix) || // slack channels have a leading #
					strings.HasPrefix(name, SweeperTargetPrefix) {
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
