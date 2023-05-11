package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAcc_TriggerResource(t *testing.T) {
	dataset := testAccDataset()

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
		Steps: []resource.TestStep{
			{
				Config: basicTriggerTestConfig(dataset),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccEnsureTriggerExists(t, "honeycombio_trigger.test"),
					resource.TestCheckResourceAttr("honeycombio_trigger.test", "name", "Test trigger from terraform-provider-honeycombio"),
					resource.TestCheckResourceAttr("honeycombio_trigger.test", "frequency", "600"),
					resource.TestCheckResourceAttr("honeycombio_trigger.test", "recipient.#", "2"),
					resource.TestCheckResourceAttrPair("honeycombio_trigger.test", "query_id", "honeycombio_query.test", "id"),
				),
			},
			{
				ResourceName:        "honeycombio_trigger.test",
				ImportStateIdPrefix: fmt.Sprintf("%v/", dataset),
				ImportState:         true,
				ImportStateVerify:   true,
			},
		},
	})
}

// TestAcc_TriggerResourceUpgradeFromVersion014 is intended to test the migration
// case from the last SDK-based version of the Trigger resource to the current Framework-based
// version.
//
// State is first generated with the SDKv2 provider and then a plan is done with the new provider to
// ensure there are no planned changes after migrating to the Framework-based resource.
//
// See: https://developer.hashicorp.com/terraform/plugin/framework/migrating/testing#testing-migration
func TestAcc_TriggerResourceUpgradeFromVersion014(t *testing.T) {
	dataset := testAccDataset()

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"honeycombio": {
						VersionConstraint: "~> 0.14",
						Source:            "honeycombio/honeycombio",
					},
				},
				Config: basicTriggerTestConfig(dataset),
				Check: resource.ComposeTestCheckFunc(
					testAccEnsureTriggerExists(t, "honeycombio_trigger.test"),
				),
			},
			{
				ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
				Config:                   basicTriggerTestConfig(dataset),
				PlanOnly:                 true,
			},
		},
	})
}

func basicTriggerTestConfig(dataset string) string {
	return fmt.Sprintf(`
data "honeycombio_query_specification" "test" {
  calculation {
    op     = "AVG"
    column = "duration_ms"
  }
  time_range = 1200
}

resource "honeycombio_query" "test" {
  dataset    = "%[1]s"
  query_json = data.honeycombio_query_specification.test.json
}

resource "honeycombio_pagerduty_recipient" "test" {
  integration_key  = "09c9d4cacd68933151a1ef1048b67dd5"
  integration_name = "acctest"
}

resource "honeycombio_trigger" "test" {
  name    = "Test trigger from terraform-provider-honeycombio"
  dataset = "%[1]s"

  description = "My nice description"

  query_id = honeycombio_query.test.id

  threshold {
    op    = ">"
    value = 100
  }

  frequency = data.honeycombio_query_specification.test.time_range / 2

  recipient {
    type   = "email"
    target = "hello@example.com"
  }

  recipient {
    id = honeycombio_pagerduty_recipient.test.id

    notification_details {
      pagerduty_severity = "info"
    }
  }
}`, dataset)
}

func testAccEnsureTriggerExists(t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resourceState, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("\"%s\" not found in state", name)
		}

		client := testAccClient(t)
		_, err := client.Triggers.Get(context.Background(), resourceState.Primary.Attributes["dataset"], resourceState.Primary.ID)
		if err != nil {
			return fmt.Errorf("failed to fetch created trigger: %w", err)
		}

		return nil
	}
}
