package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
)

func TestAcc_TriggerResource(t *testing.T) {
	dataset := testAccDataset()

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigBasicTriggerTest(dataset, "info"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccEnsureTriggerExists(t, "honeycombio_trigger.test"),
					resource.TestCheckResourceAttr("honeycombio_trigger.test", "name", "Test trigger from terraform-provider-honeycombio"),
					resource.TestCheckResourceAttr("honeycombio_trigger.test", "frequency", "600"),
					resource.TestCheckResourceAttr("honeycombio_trigger.test", "recipient.#", "2"),
					resource.TestCheckResourceAttrPair("honeycombio_trigger.test", "query_id", "honeycombio_query.test", "id"),
				),
			},
			// then update the PD Severity from info -> critical (the default)
			{
				Config: testAccConfigBasicTriggerTest(dataset, "critical"),
			},
			{
				ResourceName:        "honeycombio_trigger.test",
				ImportStateIdPrefix: fmt.Sprintf("%v/", dataset),
				ImportState:         true,
			},
		},
	})
}

// TestAcc_TriggerResourceUpgradeFromVersion014 is intended to test the migration
// case from the last SDK-based version of the Trigger resource to the current Framework-based
// version.
//
// See: https://developer.hashicorp.com/terraform/plugin/framework/migrating/testing#testing-migration
func TestAcc_TriggerResourceUpgradeFromVersion014(t *testing.T) {
	dataset := testAccDataset()

	config := testAccConfigBasicTriggerTest(dataset, "info")

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"honeycombio": {
						VersionConstraint: "~> 0.14",
						Source:            "honeycombio/honeycombio",
					},
				},
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccEnsureTriggerExists(t, "honeycombio_trigger.test"),
				),
			},
			{
				ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
				Config:                   config,
			},
		},
	})
}

func TestAcc_TriggerResourceUpdateRecipientByID(t *testing.T) {
	ctx := context.Background()
	c := testAccClient(t)
	dataset := testAccDataset()

	testRecipients := []client.Recipient{
		{
			Type: client.RecipientTypeEmail,
			Details: client.RecipientDetails{
				EmailAddress: acctest.RandString(8) + "@example.com",
			},
		},
		{
			Type: client.RecipientTypeSlack,
			Details: client.RecipientDetails{
				SlackChannel: "#" + acctest.RandString(8),
			},
		},
	}

	for i, r := range testRecipients {
		rcpt, err := c.Recipients.Create(ctx, &r)
		if err != nil {
			t.Error(err)
		}
		// update ID for removal later
		testRecipients[i].ID = rcpt.ID
	}
	t.Cleanup(func() {
		// remove recipients at the of the test run
		for _, col := range testRecipients {
			//nolint:errcheck
			c.DerivedColumns.Delete(ctx, dataset, col.ID)
		}
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigTriggerRecipientByID(dataset, testRecipients[0].ID),
				Check:  resource.TestCheckResourceAttr("honeycombio_trigger.test", "recipient.0.id", testRecipients[0].ID),
			},
			{
				Config: testAccConfigTriggerRecipientByID(dataset, testRecipients[1].ID),
				Check:  resource.TestCheckResourceAttr("honeycombio_trigger.test", "recipient.0.id", testRecipients[1].ID),
			},
		},
	})
}

func TestAcc_TriggerResourceRecipientOrderingStable(t *testing.T) {
	dataset := testAccDataset()

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "honeycombio_query_specification" "test" {
  calculation {
    op     = "COUNT"
  }

  time_range = 1200
}

resource "honeycombio_query" "test" {
  dataset    = "%[1]s"
  query_json = data.honeycombio_query_specification.test.json
}

resource "honeycombio_trigger" "test" {
  name = "Ensure mixed order recipients don't cause infinite diffs"

  query_id = honeycombio_query.test.id
  dataset  = "%[1]s"

  threshold {
    op    = ">"
    value = 1000
  }

  alert_type = "on_change"

  recipient {
    type   = "slack"
    target = "#test2"
  }

  recipient {
    type   = "slack"
    target = "#test"
  }

  recipient {
    type   = "email"
    target = "bob@example.com"
  }

  recipient {
    type   = "email"
    target = "alice@example.com"
  }

  recipient {
    type   = "marker"
    target = "trigger fired"
  }
}
`, dataset),
			},
			{
				// now remove two recipients and add a new one
				Config: fmt.Sprintf(`
data "honeycombio_query_specification" "test" {
  calculation {
    op     = "COUNT"
  }

  time_range = 1200
}

resource "honeycombio_query" "test" {
  dataset    = "%[1]s"
  query_json = data.honeycombio_query_specification.test.json
}

resource "honeycombio_trigger" "test" {
  name = "Ensure mixed order recipients don't cause infinite diffs"

  query_id = honeycombio_query.test.id
  dataset  = "%[1]s"

  threshold {
    op    = ">"
    value = 1000
  }

  alert_type = "on_change"

  recipient {
    type   = "slack"
    target = "#test"
  }

  recipient {
    type   = "email"
    target = "bob@example.com"
  }

  recipient {
    type   = "email"
    target = "alice@example.com"
  }

  recipient {
    type   = "slack"
    target = "#a-new-channel"
  }
}
`, dataset),
			},
		},
	})
}

func TestAcc_TriggerResourceEvaluationWindow(t *testing.T) {
	dataset := testAccDataset()

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "honeycombio_query_specification" "test" {
  calculation {
  op     = "COUNT"
}

  time_range = 1200
}

resource "honeycombio_query" "test" {
  dataset    = "%[1]s"
  query_json = data.honeycombio_query_specification.test.json
}

resource "honeycombio_trigger" "test" {
  name = "Ensure evaluation windows are set correctly"

  query_id = honeycombio_query.test.id
  dataset  = "%[1]s"

  threshold {
    op    = ">"
    value = 1000
  }

  evaluation_schedule {
    start_time = "13:00"
    end_time   = "21:00"

    days_of_week = ["monday", "wednesday", "friday"]
  }
}`, dataset),
			},
			{
				// update the schedule with different days and hours
				Config: fmt.Sprintf(`
data "honeycombio_query_specification" "test" {
  calculation {
  op     = "COUNT"
}

  time_range = 1200
}

resource "honeycombio_query" "test" {
  dataset    = "%[1]s"
  query_json = data.honeycombio_query_specification.test.json
}

resource "honeycombio_trigger" "test" {
  name = "Ensure evaluation windows are set correctly"

  query_id = honeycombio_query.test.id
  dataset  = "%[1]s"

  threshold {
    op    = ">"
    value = 1000
  }

  evaluation_schedule {
    start_time = "11:00"
    end_time   = "22:00"

    days_of_week = ["monday", "wednesday", "friday", "sunday"]
  }
}`, dataset),
			},
			{
				// remove the evaluation schedule to switch back to frequency
				Config: fmt.Sprintf(`
data "honeycombio_query_specification" "test" {
  calculation {
  op     = "COUNT"
}

  time_range = 1200
}

resource "honeycombio_query" "test" {
  dataset    = "%[1]s"
  query_json = data.honeycombio_query_specification.test.json
}

resource "honeycombio_trigger" "test" {
  name = "Ensure evaluation windows are set correctly"

  query_id = honeycombio_query.test.id
  dataset  = "%[1]s"

  threshold {
    op    = ">"
    value = 1000
  }
}`, dataset),
			},
		},
	})
}

func TestAcc_TriggerResourcePagerDutyUnsetSeverity(t *testing.T) {
	t.Skip("known issue: see https://github.com/honeycombio/terraform-provider-honeycombio/issues/309")

	dataset := testAccDataset()

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
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
  integration_name = "severity-test"
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
    id = honeycombio_pagerduty_recipient.test.id
    // default severity is 'critical'
  }
}`, dataset),
			},
		},
	})
}

func testAccConfigBasicTriggerTest(dataset, pdseverity string) string {
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
  integration_key  = "08b9d4cacd68933151a1ef1028b67da2"
  integration_name = "testacc-basic"
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
      pagerduty_severity = "%[2]s"
    }
  }
}`, dataset, pdseverity)
}

func testAccConfigTriggerRecipientByID(dataset, recipientID string) string {
	return fmt.Sprintf(`
data "honeycombio_query_specification" "test" {
  calculation {
    op     = "AVG"
    column = "duration_ms"
  }

  time_range = 1800
}

resource "honeycombio_query" "test" {
  dataset    = "%[1]s"
  query_json = data.honeycombio_query_specification.test.json
}

resource "honeycombio_trigger" "test" {
  name    = "Test trigger with changing recipient id"
  dataset = "%[1]s"
  query_id = honeycombio_query.test.id
  threshold {
    op    = ">"
    value = 100
  }

  recipient {
    id = "%[2]s"
  }
}`, dataset, recipientID)
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
