package provider

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
)

func TestAcc_BurnAlertResource(t *testing.T) {
	dataset, sloID := burnAlertAccTestSetup(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigBasicBurnAlertTest(dataset, sloID, "info"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccEnsureBurnAlertExists(t, "honeycombio_burn_alert.test"),
					resource.TestCheckResourceAttr("honeycombio_burn_alert.test", "slo_id", sloID),
					resource.TestCheckResourceAttr("honeycombio_burn_alert.test", "exhaustion_minutes", "240"),
					resource.TestCheckResourceAttr("honeycombio_burn_alert.test", "recipient.#", "1"),
				),
			},
			// then update the PD Severity from info -> critical (the default)
			{
				Config: testAccConfigBasicBurnAlertTest(dataset, sloID, "critical"),
			},
			{
				ResourceName:        "honeycombio_burn_alert.test",
				ImportStateIdPrefix: fmt.Sprintf("%v/", dataset),
				ImportState:         true,
			},
		},
	})
}

// TestAcc_BurnAlertResourceUpgradeFromVersion015 is intended to test the migration
// case from the last SDK-based version of the Burn Alert resource to the current Framework-based
// version.
//
// See: https://developer.hashicorp.com/terraform/plugin/framework/migrating/testing#testing-migration
func TestAcc_BurnAlertResourceUpgradeFromVersion015(t *testing.T) {
	dataset, sloID := burnAlertAccTestSetup(t)

	config := testAccConfigBasicBurnAlertTest(dataset, sloID, "info")

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"honeycombio": {
						VersionConstraint: "~> 0.15.0",
						Source:            "honeycombio/honeycombio",
					},
				},
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccEnsureBurnAlertExists(t, "honeycombio_burn_alert.test"),
				),
				SkipFunc: func() (bool, error) {
					apiHost := os.Getenv(client.DefaultAPIHostEnv)
					if apiHost == "" {
						return false, nil
					}
					return apiHost != client.DefaultAPIHost, nil
				},
			},
			{
				ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
				Config:                   config,
			},
		},
	})
}

func testAccEnsureBurnAlertExists(t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resourceState, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("\"%s\" not found in state", name)
		}

		client := testAccClient(t)
		_, err := client.BurnAlerts.Get(context.Background(), resourceState.Primary.Attributes["dataset"], resourceState.Primary.ID)
		if err != nil {
			return fmt.Errorf("failed to fetch created Burn Alert: %w", err)
		}

		return nil
	}
}

func burnAlertAccTestSetup(t *testing.T) (string, string) {
	t.Helper()

	ctx := context.Background()
	dataset := testAccDataset()
	c := testAccClient(t)

	sli, err := c.DerivedColumns.Create(ctx, dataset, &client.DerivedColumn{
		Alias:      "sli." + acctest.RandString(8),
		Expression: "BOOL(1)",
	})
	if err != nil {
		t.Error(err)
	}
	slo, err := c.SLOs.Create(ctx, dataset, &client.SLO{
		Name:             acctest.RandString(8) + " SLO",
		TimePeriodDays:   14,
		TargetPerMillion: 995000,
		SLI:              client.SLIRef{Alias: sli.Alias},
	})
	require.NoError(t, err)
	//nolint:errcheck
	t.Cleanup(func() {
		// remove SLO, SLI DC at end of test run
		c.SLOs.Delete(ctx, dataset, slo.ID)
		c.DerivedColumns.Delete(ctx, dataset, sli.ID)
	})

	return dataset, slo.ID
}

func testAccConfigBasicBurnAlertTest(dataset, sloID, pdseverity string) string {
	return fmt.Sprintf(`
resource "honeycombio_pagerduty_recipient" "test" {
  integration_key  = "08b9d4cacd68933151a1ef1028b67da2"
  integration_name = "testacc-basic"
}

resource "honeycombio_burn_alert" "test" {
  dataset            = "%[1]s"
  slo_id             = "%[2]s"
  exhaustion_minutes = 4 * 60

  recipient {
    id = honeycombio_pagerduty_recipient.test.id

    notification_details {
      pagerduty_severity = "%[3]s"
    }
  }
}`, dataset, sloID, pdseverity)
}
