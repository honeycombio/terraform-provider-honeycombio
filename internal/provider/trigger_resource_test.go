package provider

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/test"
)

func TestAcc_TriggerResource(t *testing.T) {
	dataset := testAccDataset()
	name := test.RandomStringWithPrefix("test.", 20)

	t.Run("trigger resource with Query ID", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:                 testAccPreCheck(t),
			ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
			Steps: []resource.TestStep{
				{
					Config: testAccConfigBasicTriggerTest(dataset, name, "info"),
					Check: resource.ComposeAggregateTestCheckFunc(
						testAccEnsureTriggerExists(t, "honeycombio_trigger.test"),
						resource.TestCheckResourceAttr("honeycombio_trigger.test", "name", name),
						resource.TestCheckResourceAttr("honeycombio_trigger.test", "frequency", "600"),
						resource.TestCheckResourceAttr("honeycombio_trigger.test", "recipient.#", "2"),
						resource.TestCheckResourceAttr("honeycombio_trigger.test", "threshold.0.exceeded_limit", "1"),
						resource.TestCheckResourceAttrPair("honeycombio_trigger.test", "query_id", "honeycombio_query.test", "id"),
						resource.TestCheckNoResourceAttr("honeycombio_trigger.test", "query_json"),
					),
				},
				// then update the PD Severity from info -> critical (the default)
				{
					Config: testAccConfigBasicTriggerTest(dataset, name, "critical"),
				},
				{
					ResourceName:        "honeycombio_trigger.test",
					ImportStateIdPrefix: fmt.Sprintf("%v/", dataset),
					ImportState:         true,
				},
			},
		})
	})

	t.Run("trigger resource with Query JSON", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:                 testAccPreCheck(t),
			ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
			Steps: []resource.TestStep{
				{
					Config: testAccConfigBasicTriggerTest_QuerySpec(dataset, name, "info"),
					Check: resource.ComposeAggregateTestCheckFunc(
						testAccEnsureTriggerExists(t, "honeycombio_trigger.test"),
						resource.TestCheckResourceAttr("honeycombio_trigger.test", "name", name),
						resource.TestCheckResourceAttr("honeycombio_trigger.test", "frequency", "600"),
						resource.TestCheckResourceAttr("honeycombio_trigger.test", "recipient.#", "2"),
						resource.TestCheckResourceAttr("honeycombio_trigger.test", "threshold.0.exceeded_limit", "1"),
						resource.TestCheckNoResourceAttr("honeycombio_trigger.test", "query_id"),
					),
				},
				// then update the PD Severity from info -> critical (the default)
				{
					Config: testAccConfigBasicTriggerTest_QuerySpec(dataset, name, "critical"),
				},
				{
					ResourceName:        "honeycombio_trigger.test",
					ImportStateIdPrefix: fmt.Sprintf("%v/", dataset),
					ImportState:         true,
				},
			},
		})
	})

	t.Run("trigger resource with custom webhook recipient", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:                 testAccPreCheck(t),
			ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
			Steps: []resource.TestStep{
				{
					Config: testAccConfigBasicTriggerTestWithWebhookRecip(dataset, name, "info"),
					Check: resource.ComposeAggregateTestCheckFunc(
						testAccEnsureTriggerExists(t, "honeycombio_trigger.test"),
						resource.TestCheckResourceAttr("honeycombio_trigger.test", "name", name),
						resource.TestCheckResourceAttr("honeycombio_trigger.test", "frequency", "600"),
						resource.TestCheckResourceAttr("honeycombio_trigger.test", "recipient.#", "1"),
						resource.TestCheckResourceAttr("honeycombio_trigger.test", "recipient.0.notification_details.#", "1"),
						resource.TestCheckResourceAttr("honeycombio_trigger.test", "recipient.0.notification_details.0.variable.#", "1"),
						resource.TestCheckResourceAttr("honeycombio_trigger.test", "recipient.0.notification_details.0.variable.0.name", "severity"),
						resource.TestCheckResourceAttr("honeycombio_trigger.test", "recipient.0.notification_details.0.variable.0.value", "info"),
						resource.TestCheckResourceAttr("honeycombio_trigger.test", "threshold.0.exceeded_limit", "1"),
						resource.TestCheckResourceAttrPair("honeycombio_trigger.test", "query_id", "honeycombio_query.test", "id"),
						resource.TestCheckNoResourceAttr("honeycombio_trigger.test", "query_json"),
					),
				},
				// then update the variable value from info -> critical
				{
					Config: testAccConfigBasicTriggerTestWithWebhookRecip(dataset, name, "critical"),
					Check: resource.ComposeAggregateTestCheckFunc(
						testAccEnsureTriggerExists(t, "honeycombio_trigger.test"),
						resource.TestCheckResourceAttr("honeycombio_trigger.test", "name", name),
						resource.TestCheckResourceAttr("honeycombio_trigger.test", "frequency", "600"),
						resource.TestCheckResourceAttr("honeycombio_trigger.test", "recipient.#", "1"),
						resource.TestCheckResourceAttr("honeycombio_trigger.test", "recipient.0.notification_details.#", "1"),
						resource.TestCheckResourceAttr("honeycombio_trigger.test", "recipient.0.notification_details.0.variable.#", "1"),
						resource.TestCheckResourceAttr("honeycombio_trigger.test", "recipient.0.notification_details.0.variable.0.name", "severity"),
						resource.TestCheckResourceAttr("honeycombio_trigger.test", "recipient.0.notification_details.0.variable.0.value", "critical"),
						resource.TestCheckResourceAttr("honeycombio_trigger.test", "threshold.0.exceeded_limit", "1"),
						resource.TestCheckResourceAttrPair("honeycombio_trigger.test", "query_id", "honeycombio_query.test", "id"),
						resource.TestCheckNoResourceAttr("honeycombio_trigger.test", "query_json"),
					),
				},
				// remove variables
				{
					Config: testAccConfigBasicTriggerTestWithWebhookRecip(dataset, name, ""),
					Check: resource.ComposeAggregateTestCheckFunc(
						testAccEnsureTriggerExists(t, "honeycombio_trigger.test"),
						resource.TestCheckResourceAttr("honeycombio_trigger.test", "name", name),
						resource.TestCheckResourceAttr("honeycombio_trigger.test", "frequency", "600"),
						resource.TestCheckResourceAttr("honeycombio_trigger.test", "recipient.#", "1"),
						resource.TestCheckResourceAttr("honeycombio_trigger.test", "recipient.0.notification_details.#", "0"),
						resource.TestCheckResourceAttr("honeycombio_trigger.test", "threshold.0.exceeded_limit", "1"),
						resource.TestCheckResourceAttrPair("honeycombio_trigger.test", "query_id", "honeycombio_query.test", "id"),
						resource.TestCheckNoResourceAttr("honeycombio_trigger.test", "query_json"),
					),
				},
				{
					ResourceName:        "honeycombio_trigger.test",
					ImportStateIdPrefix: fmt.Sprintf("%v/", dataset),
					ImportState:         true,
				},
			},
		})
	})

	t.Run("duplicate variable on custom webhook recipient", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:                 testAccPreCheck(t),
			ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
			Steps: []resource.TestStep{
				{
					Config:      testAccConfigBasicTriggerTestWithWebhookRecipAndDuplicateVar(dataset, name),
					ExpectError: regexp.MustCompile(`cannot have more than one "variable" with the same "name"`),
				},
			},
		})
	})

	t.Run("trigger resource with baseline_details", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:                 testAccPreCheck(t),
			ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
			Steps: []resource.TestStep{
				// create a trigger with baseline_details
				{
					Config: testAccConfigBasicTriggerWithBaselineDetailsTest(dataset, name, client.TriggerBaselineDetails{Type: "value", OffsetMinutes: 1440}, false),
					Check: resource.ComposeAggregateTestCheckFunc(
						testAccEnsureTriggerExists(t, "honeycombio_trigger.test"),
						resource.TestCheckResourceAttr("honeycombio_trigger.test", "name", name),
						resource.TestCheckResourceAttr("honeycombio_trigger.test", "frequency", "1200"),
						resource.TestCheckResourceAttr("honeycombio_trigger.test", "recipient.#", "2"),
						resource.TestCheckResourceAttr("honeycombio_trigger.test", "threshold.0.exceeded_limit", "1"),
						resource.TestCheckResourceAttrPair("honeycombio_trigger.test", "query_id", "honeycombio_query.test", "id"),
						resource.TestCheckNoResourceAttr("honeycombio_trigger.test", "query_json"),
						resource.TestCheckResourceAttr("honeycombio_trigger.test", "baseline_details.0.offset_minutes", "1440"),
					),
				},
				// update with no baseline_details (""), baseline_details should stay the same
				{
					Config: testAccConfigBasicTriggerWithBaselineDetailsTest(dataset, name, client.TriggerBaselineDetails{}, false),
					Check: resource.ComposeAggregateTestCheckFunc(
						testAccEnsureTriggerExists(t, "honeycombio_trigger.test"),
						resource.TestCheckNoResourceAttr("honeycombio_trigger.test", "query_json"),
						resource.TestCheckResourceAttr("honeycombio_trigger.test", "baseline_details.0.offset_minutes", "1440"),
					),
				},
				// // update with empty baseline_details object (baseline_details {} ), baseline_details should be removed from trigger
				{
					Config: testAccConfigBasicTriggerWithBaselineDetailsTest(dataset, name, client.TriggerBaselineDetails{}, true),
					Check: resource.ComposeAggregateTestCheckFunc(
						testAccEnsureTriggerExists(t, "honeycombio_trigger.test"),
						resource.TestCheckNoResourceAttr("honeycombio_trigger.test", "baseline_details.#"),
					),
				},
				{
					ResourceName:        "honeycombio_trigger.test",
					ImportStateIdPrefix: fmt.Sprintf("%v/", dataset),
					ImportState:         true,
				},
			},
		})
	})
}

// TestAcc_TriggerResourceUpgradeFromVersion014 is intended to test the migration
// case from the last SDK-based version of the Trigger resource to the current Framework-based
// version.
//
// See: https://developer.hashicorp.com/terraform/plugin/framework/migrating/testing#testing-migration
func TestAcc_TriggerResourceUpgradeFromVersion014(t *testing.T) {
	dataset := testAccDataset()

	config := testAccConfigBasicTriggerTest(dataset, test.RandomStringWithPrefix("test.", 20), "info")

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"honeycombio": {
						VersionConstraint: "~> 0.14.0",
						Source:            "honeycombio/honeycombio",
					},
				},
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccEnsureTriggerExists(t, "honeycombio_trigger.test"),
				),
				// These tests pull in older versions of the provider that don't
				// support setting the API host easily. We'll skip them for now
				// if we have a non-default API host.
				SkipFunc: func() (bool, error) {
					apiHost := os.Getenv(client.DefaultAPIEndpointEnv)
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

func TestAcc_TriggerResourceUpdateRecipientByID(t *testing.T) {
	ctx := context.Background()
	c := testAccClient(t)
	dataset := testAccDataset()
	name := test.RandomStringWithPrefix("test.", 20)

	testRecipients := []client.Recipient{
		{
			Type: client.RecipientTypeEmail,
			Details: client.RecipientDetails{
				EmailAddress: test.RandomEmail(),
			},
		},
		{
			Type: client.RecipientTypeSlack,
			Details: client.RecipientDetails{
				SlackChannel: test.RandomStringWithPrefix("#test.", 8),
			},
		},
	}

	for i, r := range testRecipients {
		rcpt, err := c.Recipients.Create(ctx, &r)
		require.NoError(t, err)
		// update ID for removal later
		testRecipients[i].ID = rcpt.ID
	}
	t.Cleanup(func() {
		// remove recipients at the of the test run
		for _, col := range testRecipients {
			c.DerivedColumns.Delete(ctx, dataset, col.ID)
		}
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigTriggerRecipientByID(dataset, name, testRecipients[0].ID, 3),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("honeycombio_trigger.test", "recipient.0.id", testRecipients[0].ID),
					resource.TestCheckResourceAttr("honeycombio_trigger.test", "threshold.0.exceeded_limit", "3"),
				),
			},
			{
				Config: testAccConfigTriggerRecipientByID(dataset, name, testRecipients[1].ID, 5),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("honeycombio_trigger.test", "recipient.0.id", testRecipients[1].ID),
					resource.TestCheckResourceAttr("honeycombio_trigger.test", "threshold.0.exceeded_limit", "5"),
				),
			},
		},
	})
}

func TestAcc_TriggerResourceRecipientOrderingStable(t *testing.T) {
	dataset := testAccDataset()
	email1 := test.RandomEmail()
	email2 := test.RandomEmail()
	slack1 := test.RandomStringWithPrefix("#test.", 8)
	slack2 := test.RandomStringWithPrefix("#test.", 8)

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
    target = "%[2]s"
  }

  recipient {
    type   = "slack"
    target = "%[3]s"
  }

  recipient {
    type   = "email"
    target = "%[4]s"
  }

  recipient {
    type   = "email"
    target = "%[5]s"
  }

  recipient {
    type   = "marker"
    target = "trigger fired"
  }
}
`, dataset, slack1, slack2, email1, email2),
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
    target = "%[2]s"
  }

  recipient {
    type   = "email"
    target = "%[3]s"
  }

  recipient {
    type   = "email"
    target = "%[4]s"
  }

  recipient {
    type   = "slack"
    target = "%[5]s"
  }
}
`, dataset, slack1, email2, email1, test.RandomStringWithPrefix("#test.", 8)),
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

	t.Run("handles dynamic evaluation schedule block", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:                 testAccPreCheck(t),
			ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
			Steps: []resource.TestStep{
				{
					Config: fmt.Sprintf(`
locals {
  business_hour = {
    start_time = "14:00"
    end_time   = "22:00"
    days_of_week = [
      "monday",
      "tuesday",
      "wednesday",
      "thursday",
      "friday",
    ]
  }
}

variable "enable_evaluation_schedule_business_hour" {
  type    = bool
  default = true
}

data "honeycombio_query_specification" "test" {}

resource "honeycombio_trigger" "test" {
  name     = "test"
  dataset  = "%s"

  query_json = data.honeycombio_query_specification.test.json

  frequency = 1800

  threshold {
    exceeded_limit = 1
    op             = ">"
    value          = "0"
  }

  dynamic "evaluation_schedule" {
    for_each = var.enable_evaluation_schedule_business_hour ? [true] : []

    content {
      start_time   = local.business_hour.start_time
      end_time     = local.business_hour.end_time
      days_of_week = local.business_hour.days_of_week
    }
  }
}`, testAccDataset()),
				},
			},
		})
	})
}

func TestAcc_TriggerResourcePagerDutyUnsetSeverity(t *testing.T) {
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
    value = 1 - 0.99
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

func TestAcc_TriggerResourceHandlesRecipientChangedOutsideOfTerraform(t *testing.T) {
	c := testAccClient(t)
	ctx := context.Background()
	dataset := testAccDataset()

	// setup a slack recipient to be used in the trigger, and modified outside of terraform
	channel := test.RandomStringWithPrefix("#test.", 8)
	rcpt, err := c.Recipients.Create(ctx, &client.Recipient{
		Type: client.RecipientTypeSlack,
		Details: client.RecipientDetails{
			SlackChannel: channel,
		},
	})
	require.NoError(t, err, "failed to create test recipient")
	t.Cleanup(func() {
		c.Recipients.Delete(ctx, rcpt.ID)
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigTriggerWithSlackRecipient(dataset, channel),
			},
			{
				PreConfig: func() {
					// update the channel name outside of Terraform
					channel += "-1"
					rcpt.Details.SlackChannel = channel
					_, err := c.Recipients.Update(ctx, rcpt)
					require.NoError(t, err, "failed to update test recipient")
				},
				Config: testAccConfigTriggerWithSlackRecipient(dataset, channel),
			},
		},
	})
}

func TestAcc_TriggerResourceValidatesQueryJSON(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
		Steps: []resource.TestStep{
			{
				Config: `
resource "honeycombio_trigger" "test" {
  name    = "I fail validation"
  dataset = "foobar"

  query_id   = "jsdfjsdf"
  query_json = "{}"
}`,
				ExpectError: regexp.MustCompile(`"query_id" cannot be specified when "query_json" is specified`),
			},
			{
				Config: `
data "honeycombio_query_specification" "test" {
  calculation {
    op = "COUNT"
  }

  calculation {
    op = "AVG"
    column = "duration_ms"
  }
}

resource "honeycombio_trigger" "test" {
  name    = "I fail validation"
  dataset = "foobar"

  threshold {
    op    = ">"
    value = 100
  }

  query_json = data.honeycombio_query_specification.test.json
}`,
				ExpectError: regexp.MustCompile(`queries must contain a single calculation`),
			},
			{
				Config: `
data "honeycombio_query_specification" "test" {
  calculation {
    op = "HEATMAP"
    column = "duration_ms"
  }
}

resource "honeycombio_trigger" "test" {
  name    = "I fail validation"
  dataset = "foobar"

  threshold {
    op    = ">"
    value = 100
  }

  query_json = data.honeycombio_query_specification.test.json
}`,
				ExpectError: regexp.MustCompile(`queries cannot use HEATMAP calculations`),
			},
			{
				Config: `
data "honeycombio_query_specification" "test" {
  calculation {
    op = "CONCURRENCY"
  }
}

resource "honeycombio_trigger" "test" {
  name    = "I fail validation"
  dataset = "foobar"

  threshold {
    op    = ">"
    value = 100
  }

  query_json = data.honeycombio_query_specification.test.json
}`,
				ExpectError: regexp.MustCompile(`queries cannot use CONCURRENCY calculations`),
			},
			{
				Config: `
data "honeycombio_query_specification" "test" {
  calculation {
    op = "COUNT"
  }

  order {
    op    = "COUNT"
    order = "ascending"
  }
}

resource "honeycombio_trigger" "test" {
  name    = "I fail validation"
  dataset = "foobar"

  threshold {
    op    = ">"
    value = 100
  }

  query_json = data.honeycombio_query_specification.test.json
}`,
				ExpectError: regexp.MustCompile(`queries cannot use orders`),
			},
			{
				Config: `
data "honeycombio_query_specification" "test" {
  calculation {
    op = "COUNT"
  }

  limit = 10
}

resource "honeycombio_trigger" "test" {
  name    = "I fail validation"
  dataset = "foobar"

  threshold {
    op    = ">"
    value = 100
  }

  query_json = data.honeycombio_query_specification.test.json
}`,
				ExpectError: regexp.MustCompile(`queries cannot use limit`),
			},
			{
				Config: `
data "honeycombio_query_specification" "test" {
  calculation {
    op = "COUNT"
  }

  end_time   = 1454808600
}

resource "honeycombio_trigger" "test" {
  name    = "I fail validation"
  dataset = "foobar"

  threshold {
    op    = ">"
    value = 100
  }

  query_json = data.honeycombio_query_specification.test.json
}`,
				ExpectError: regexp.MustCompile(`queries cannot use start_time or end_time`),
			},
			{
				Config: `
data "honeycombio_query_specification" "test" {
  calculation {
    op = "COUNT"
  }

  granularity = 120
}

resource "honeycombio_trigger" "test" {
  name    = "I fail validation"
  dataset = "foobar"

  threshold {
    op    = ">"
    value = 100
  }

  query_json = data.honeycombio_query_specification.test.json
}`,
				ExpectError: regexp.MustCompile(`queries cannot use granularity`),
			},
			{
				Config: `
data "honeycombio_query_specification" "test" {
  calculation {
    op = "COUNT"
  }

  time_range = 1800
}

resource "honeycombio_trigger" "test" {
  name    = "I fail validation"
  dataset = "foobar"

  threshold {
    op    = ">"
    value = 100
  }

  frequency = 7200

  query_json = data.honeycombio_query_specification.test.json
}`,
				ExpectError: regexp.MustCompile(`frequency must be at least equal to the query duration`),
			},
			{
				Config: `
data "honeycombio_query_specification" "test" {
  calculation {
    op = "COUNT"
  }
}

resource "honeycombio_trigger" "test" {
  name    = "I fail validation"
  dataset = "foobar"

  threshold {
    op    = ">"
    value = 100
  }

  query_json = data.honeycombio_query_specification.test.json
}`,
				ExpectError: regexp.MustCompile(`duration cannot be more than four times the frequency`),
			},
		},
	})
}

func TestAcc_TriggerResource_QueryJSONHandlesEquivQuerySpecs(t *testing.T) {
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

  filter_combination = "AND"

  time_range = 1200
}

resource honeycombio_trigger "test" {
  name    = "test trigger"
  dataset = "%s"

  threshold {
    op    = ">"
    value = 100
  }

  frequency = data.honeycombio_query_specification.test.time_range / 2

  query_json = data.honeycombio_query_specification.test.json
}`, dataset),
			},
		},
	})
}

func testAccConfigBasicTriggerTest(dataset, name, pdseverity string) string {
	email := test.RandomEmail()
	pdKey := test.RandomString(32)
	pdName := test.RandomStringWithPrefix("test.", 20)

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
  integration_key  = "%[5]s"
  integration_name = "%[6]s"
}

resource "honeycombio_trigger" "test" {
  name    = "%[2]s"
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
    target = "%[4]s"
  }

  recipient {
    id = honeycombio_pagerduty_recipient.test.id

    notification_details {
      pagerduty_severity = "%[3]s"
    }
  }
}`, dataset, name, pdseverity, email, pdKey, pdName)
}

func testAccConfigBasicTriggerWithBaselineDetailsTest(dataset string, name string, baseline_details client.TriggerBaselineDetails, omitBaselineDetails bool) string {
	email := test.RandomEmail()
	pdKey := test.RandomString(32)
	pdName := test.RandomStringWithPrefix("test.", 20)

	baselineDetails := ""

	if baseline_details.Type != "" {
		baselineDetails = fmt.Sprintf(`
  baseline_details {
    type           = "%s"
    offset_minutes = %d
  }`, baseline_details.Type, baseline_details.OffsetMinutes)
	}

	if omitBaselineDetails {
		baselineDetails = `baseline_details {}`
	}

	return fmt.Sprintf(`
data "honeycombio_query_specification" "test" {
  calculation {
    op     = "AVG"
    column = "db_dur_ms"
  }
  time_range = 1200
}

resource "honeycombio_query" "test" {
  dataset    = "%[1]s"
  query_json = data.honeycombio_query_specification.test.json
}

resource "honeycombio_pagerduty_recipient" "test" {
  integration_key  = "%[5]s"
  integration_name = "%[6]s"
}

resource "honeycombio_trigger" "test" {
  name    = "%[2]s"
  dataset = "%[1]s"

  description = "My nice description"

  query_id = honeycombio_query.test.id

  threshold {
    op    = ">="
    value = 100
  }

	frequency = 1200

  recipient {
    type   = "email"
    target = "%[4]s"
  }

  recipient {
    id = honeycombio_pagerduty_recipient.test.id

    notification_details {
      pagerduty_severity = "info"
    }
  }

  %[3]s
}`, dataset, name, baselineDetails, email, pdKey, pdName)
}

func testAccConfigBasicTriggerTestWithWebhookRecip(dataset, name, varValue string) string {
	tmplBody := `<<EOT
		{
			"name": " {{ .Name }}",
			"id": " {{ .ID }}",
			"description": " {{ .Description }}",
		}
		EOT`

	if varValue == "" {
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

resource "honeycombio_webhook_recipient" "test" {
  name = "test"
	url  = "http://example.com"

	header {
	  name = "Authorization"
	  value = "Bearer abc123"
	}

	variable {
	  name = "severity"
      default_value = "critical"
	}

	template {
	  type   = "trigger"
      body 	 = %[3]s
    }
}

resource "honeycombio_trigger" "test" {
  name    = "%[2]s"
  dataset = "%[1]s"

  description = "My nice description"

  query_id = honeycombio_query.test.id

  threshold {
    op    = ">"
    value = 100
  }

  frequency = data.honeycombio_query_specification.test.time_range / 2

  recipient {
	id = honeycombio_webhook_recipient.test.id
  }
}`, dataset, name, tmplBody)
	}

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

resource "honeycombio_webhook_recipient" "test" {
  name = "test"
	url  = "http://example.com"

	header {
	  name = "Authorization"
	  value = "Bearer abc123"
	}

	variable {
	  name = "severity"
      default_value = "critical"
	}

	template {
	  type   = "trigger"
      body 	 = %[4]s
    }
}

resource "honeycombio_trigger" "test" {
  name    = "%[2]s"
  dataset = "%[1]s"

  description = "My nice description"

  query_id = honeycombio_query.test.id

  threshold {
    op    = ">"
    value = 100
  }

  frequency = data.honeycombio_query_specification.test.time_range / 2

  recipient {
	id = honeycombio_webhook_recipient.test.id
	
	notification_details {	
		variable {
			name = "severity"
			value = "%[3]s"
		}
	}
  }
}`, dataset, name, varValue, tmplBody)
}

func testAccConfigBasicTriggerTestWithWebhookRecipAndDuplicateVar(dataset, name string) string {
	tmplBody := `<<EOT
		{
			"name": " {{ .Name }}",
			"id": " {{ .ID }}",
			"description": " {{ .Description }}",
		}
		EOT`

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

resource "honeycombio_webhook_recipient" "test" {
  name = "test"
	url  = "http://example.com"

	header {
	  name = "Authorization"
	  value = "Bearer abc123"
	}

	variable {
	  name = "severity"
      default_value = "critical"
	}

	template {
	  type   = "trigger"
      body 	 = %[3]s
    }
}

resource "honeycombio_trigger" "test" {
  name    = "%[2]s"
  dataset = "%[1]s"

  description = "My nice description"

  query_id = honeycombio_query.test.id

  threshold {
    op    = ">"
    value = 100
  }

  frequency = data.honeycombio_query_specification.test.time_range / 2

  recipient {
	id = honeycombio_webhook_recipient.test.id
	
	notification_details {	
		variable {
			name = "severity"
			value = "info"
		}

		variable {
			name = "severity"
			value = "critical"
		}
	}
  }
}`, dataset, name, tmplBody)
}

func testAccConfigBasicTriggerTest_QuerySpec(dataset, name, pdseverity string) string {
	email := test.RandomEmail()
	pdKey := test.RandomString(32)
	pdName := test.RandomStringWithPrefix("test.", 20)

	return fmt.Sprintf(`
data "honeycombio_query_specification" "test" {
  calculation {
    op     = "AVG"
    column = "duration_ms"
  }
  time_range = 1200
}

resource "honeycombio_pagerduty_recipient" "test" {
  integration_key  = "%[5]s"
  integration_name = "%[6]s"
}

resource "honeycombio_trigger" "test" {
  name    = "%[2]s"
  dataset = "%[1]s"

  description = "My nice description"

  query_json = data.honeycombio_query_specification.test.json

  threshold {
    op    = ">"
    value = 100
  }

  frequency = data.honeycombio_query_specification.test.time_range / 2

  recipient {
    type   = "email"
    target = "%[4]s"
  }

  recipient {
    id = honeycombio_pagerduty_recipient.test.id

    notification_details {
      pagerduty_severity = "%[3]s"
    }
  }
}`, dataset, name, pdseverity, email, pdKey, pdName)
}

func testAccConfigTriggerRecipientByID(dataset, name, recipientID string, exceededLimit int) string {
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
  name    = "%[2]s"
  dataset = "%[1]s"
  query_id = honeycombio_query.test.id

  threshold {
    op             = ">"
    value          = 100
    exceeded_limit = %[4]d
  }

  recipient {
    id = "%[3]s"
  }
}`, dataset, name, recipientID, exceededLimit)
}

func testAccConfigTriggerWithSlackRecipient(dataset, channel string) string {
	return fmt.Sprintf(`
data "honeycombio_query_specification" "test" {}

resource "honeycombio_query" "test" {
  dataset    = "%[1]s"
  query_json = data.honeycombio_query_specification.test.json
}

resource "honeycombio_trigger" "test" {
  name = "my test trigger"

  dataset  = "%[1]s"
  query_id = honeycombio_query.test.id

  threshold {
    op    = "<"
    value = 100
  }

  frequency = 1800

  recipient {
    type   = "slack"
    target = "%[2]s"
  }
}`, dataset, channel)
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
