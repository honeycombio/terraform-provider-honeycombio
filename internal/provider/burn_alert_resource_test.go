package provider

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"

	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/test"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
)

const testBADescription = "burn alert description"

func TestAcc_BurnAlertResource_defaultBasic(t *testing.T) {
	dataset, sloID := burnAlertAccTestSetup(t)
	burnAlert := &client.BurnAlert{}

	// Create
	exhaustionMinutes := 240

	// Update
	updatedExhaustionMinutes := 480
	budgetRateWindowMinutes := 60
	budgetRateDecreasePercent := 0.0001

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
		CheckDestroy:             testAccEnsureBurnAlertDestroyed(t),
		Steps: []resource.TestStep{
			// Create - basic
			{
				Config: testAccConfigBurnAlertDefault_basic(exhaustionMinutes, dataset, sloID, "info"),
				Check:  testAccEnsureSuccessExhaustionTimeAlert(t, burnAlert, exhaustionMinutes, "info", sloID),
			},
			// Update - PD Severity from info -> critical (the default)
			{
				Config: testAccConfigBurnAlertDefault_basic(exhaustionMinutes, dataset, sloID, "critical"),
				Check:  testAccEnsureSuccessExhaustionTimeAlert(t, burnAlert, exhaustionMinutes, "critical", sloID),
			},
			// Import
			{
				ResourceName:            "honeycombio_burn_alert.test",
				ImportStateIdPrefix:     fmt.Sprintf("%v/", dataset),
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"recipient"},
			},
			// Update - exhaustion time to exhaustion time
			{
				Config: testAccConfigBurnAlertDefault_basic(updatedExhaustionMinutes, dataset, sloID, "info"),
				Check:  testAccEnsureSuccessExhaustionTimeAlert(t, burnAlert, updatedExhaustionMinutes, "info", sloID),
			},
			// Update - exhaustion time to budget rate
			{
				Config: testAccConfigBurnAlertBudgetRate_basic(budgetRateWindowMinutes, budgetRateDecreasePercent, dataset, sloID, "info"),
				Check:  testAccEnsureSuccessBudgetRateAlert(t, burnAlert, budgetRateWindowMinutes, budgetRateDecreasePercent, "info", sloID),
			},
		},
	})
}

func TestAcc_BurnAlertResource_exhaustionTimeBasic(t *testing.T) {
	dataset, sloID := burnAlertAccTestSetup(t)
	burnAlert := &client.BurnAlert{}

	// Create
	exhaustionMinutes := 0

	// Update
	updatedExhaustionMinutes := 240
	budgetRateWindowMinutes := 60
	budgetRateDecreasePercent := 5.1

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
		CheckDestroy:             testAccEnsureBurnAlertDestroyed(t),
		Steps: []resource.TestStep{
			// Create - basic
			{
				Config: testAccConfigBurnAlertDefault_basic(exhaustionMinutes, dataset, sloID, "info"),
				Check:  testAccEnsureSuccessExhaustionTimeAlert(t, burnAlert, exhaustionMinutes, "info", sloID),
			},
			// Update - PD Severity from info -> critical (the default)
			{
				Config: testAccConfigBurnAlertDefault_basic(exhaustionMinutes, dataset, sloID, "critical"),
				Check:  testAccEnsureSuccessExhaustionTimeAlert(t, burnAlert, exhaustionMinutes, "critical", sloID),
			},
			// Import
			{
				ResourceName:            "honeycombio_burn_alert.test",
				ImportStateIdPrefix:     fmt.Sprintf("%v/", dataset),
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"recipient"},
			},
			// Update - exhaustion time to exhaustion time
			{
				Config: testAccConfigBurnAlertDefault_basic(updatedExhaustionMinutes, dataset, sloID, "info"),
				Check:  testAccEnsureSuccessExhaustionTimeAlert(t, burnAlert, updatedExhaustionMinutes, "info", sloID),
			},
			// Update - exhaustion time to budget rate
			{
				Config: testAccConfigBurnAlertBudgetRate_basic(budgetRateWindowMinutes, budgetRateDecreasePercent, dataset, sloID, "info"),
				Check:  testAccEnsureSuccessBudgetRateAlert(t, burnAlert, budgetRateWindowMinutes, budgetRateDecreasePercent, "info", sloID),
			},
		},
	})
}

func TestAcc_BurnAlertResource_exhaustionTimeBasicWebhookRecipient(t *testing.T) {
	dataset, sloID := burnAlertAccTestSetup(t)
	burnAlert := &client.BurnAlert{}
	exhaustionMinutes := 240

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
		CheckDestroy:             testAccEnsureBurnAlertDestroyed(t),
		Steps: []resource.TestStep{
			// Create - basic
			{
				Config: testAccConfigBurnAlertExhaustionTime_basicWebhookRecipient(exhaustionMinutes, dataset, sloID, "warning"),
				Check:  testAccEnsureSuccessExhaustionTimeAlertWithWebhookRecip(t, burnAlert, exhaustionMinutes, sloID, "warning"),
			},
			// Update - change variable value
			{
				Config: testAccConfigBurnAlertExhaustionTime_basicWebhookRecipient(exhaustionMinutes, dataset, sloID, "info"),
				Check:  testAccEnsureSuccessExhaustionTimeAlertWithWebhookRecip(t, burnAlert, exhaustionMinutes, sloID, "info"),
			},
			// Import
			{
				ResourceName:            "honeycombio_burn_alert.test",
				ImportStateIdPrefix:     fmt.Sprintf("%v/", dataset),
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"recipient"},
			},
		},
	})
}

func TestAcc_BurnAlertResource_budgetRateBasic(t *testing.T) {
	dataset, sloID := burnAlertAccTestSetup(t)
	burnAlert := &client.BurnAlert{}

	// Create
	budgetRateWindowMinutes := 60
	budgetRateDecreasePercent := float64(5)

	// Update
	updatedBudgetRateWindowMinutes := 100
	updatedBudgetRateDecreasePercent := 0.1
	exhaustionTime := 0

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
		CheckDestroy:             testAccEnsureBurnAlertDestroyed(t),
		Steps: []resource.TestStep{
			// Create - basic
			{
				Config: testAccConfigBurnAlertBudgetRate_basic(budgetRateWindowMinutes, budgetRateDecreasePercent, dataset, sloID, "info"),
				Check:  testAccEnsureSuccessBudgetRateAlert(t, burnAlert, budgetRateWindowMinutes, budgetRateDecreasePercent, "info", sloID),
			},
			// Update - PD Severity from info -> critical (the default)
			{
				Config: testAccConfigBurnAlertBudgetRate_basic(budgetRateWindowMinutes, budgetRateDecreasePercent, dataset, sloID, "critical"),
				Check:  testAccEnsureSuccessBudgetRateAlert(t, burnAlert, budgetRateWindowMinutes, budgetRateDecreasePercent, "critical", sloID),
			},
			// Import
			{
				ResourceName:            "honeycombio_burn_alert.test",
				ImportStateIdPrefix:     fmt.Sprintf("%v/", dataset),
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"recipient"},
			},
			// Update - budget rate to budget rate
			{
				Config: testAccConfigBurnAlertBudgetRate_basic(updatedBudgetRateWindowMinutes, updatedBudgetRateDecreasePercent, dataset, sloID, "info"),
				Check:  testAccEnsureSuccessBudgetRateAlert(t, burnAlert, updatedBudgetRateWindowMinutes, updatedBudgetRateDecreasePercent, "info", sloID),
			},
			// Update - budget rate to exhaustion time
			{
				Config: testAccConfigBurnAlertExhaustionTime_basic(exhaustionTime, dataset, sloID, "info"),
				Check:  testAccEnsureSuccessExhaustionTimeAlert(t, burnAlert, exhaustionTime, "info", sloID),
			},
		},
	})
}

func TestAcc_BurnAlertResource_budgetRateBasicWebhookRecipient(t *testing.T) {
	dataset, sloID := burnAlertAccTestSetup(t)
	burnAlert := &client.BurnAlert{}
	budgetRateWindowMinutes := 60
	budgetRateDecreasePercent := float64(5)

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
		CheckDestroy:             testAccEnsureBurnAlertDestroyed(t),
		Steps: []resource.TestStep{
			// Create - basic
			{
				Config: testAccConfigBurnAlertBudgetRate_basicWebhookRecipient(budgetRateWindowMinutes, budgetRateDecreasePercent, dataset, sloID, "warning"),
				Check:  testAccEnsureSuccessBudgetRateAlertWithWebhookRecip(t, burnAlert, budgetRateWindowMinutes, budgetRateDecreasePercent, sloID, "warning"),
			},
			// Update - change variable value
			{
				Config: testAccConfigBurnAlertBudgetRate_basicWebhookRecipient(budgetRateWindowMinutes, budgetRateDecreasePercent, dataset, sloID, "info"),
				Check:  testAccEnsureSuccessBudgetRateAlertWithWebhookRecip(t, burnAlert, budgetRateWindowMinutes, budgetRateDecreasePercent, sloID, "info"),
			},
			// Import
			{
				ResourceName:            "honeycombio_burn_alert.test",
				ImportStateIdPrefix:     fmt.Sprintf("%v/", dataset),
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"recipient"},
			},
		},
	})
}

// Check that creating a budget rate alert with a
// budget_rate_decrease_percent with trailing zeros works,
// doesn't produce spurious plans after apply, and imports successfully
func TestAcc_BurnAlertResource_budgetRateTrailingZeros(t *testing.T) {
	dataset, sloID := burnAlertAccTestSetup(t)
	burnAlert := &client.BurnAlert{}

	// Create
	budgetRateWindowMinutes := 60
	budgetRateDecreasePercent := float64(5.00000)
	pdSeverity := "info"

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
		CheckDestroy:             testAccEnsureBurnAlertDestroyed(t),
		Steps: []resource.TestStep{
			// Create - trailing zeros on budget_rate_decrease_percent
			{
				Config: testAccConfigBurnAlertBudgetRate_trailingZeros(dataset, sloID),
				Check:  testAccEnsureSuccessBudgetRateAlert(t, burnAlert, budgetRateWindowMinutes, budgetRateDecreasePercent, pdSeverity, sloID),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
			// Import
			{
				ResourceName:            "honeycombio_burn_alert.test",
				ImportStateIdPrefix:     fmt.Sprintf("%v/", dataset),
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"recipient"},
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
	burnAlert := &client.BurnAlert{}

	config := testAccConfigBurnAlert_withoutDescription(60, dataset, sloID, "info")

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
					testAccEnsureBurnAlertExists(t, "honeycombio_burn_alert.test", burnAlert),
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

func TestAcc_BurnAlertResource_Import_validateImportID(t *testing.T) {
	dataset, sloID := burnAlertAccTestSetup(t)
	burnAlert := &client.BurnAlert{}

	exhaustionMinutes := 240

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
		CheckDestroy:             testAccEnsureBurnAlertDestroyed(t),
		Steps: []resource.TestStep{
			// Create resource for importing
			{
				Config: testAccConfigBurnAlertDefault_basic(exhaustionMinutes, dataset, sloID, "info"),
				Check:  testAccEnsureSuccessExhaustionTimeAlert(t, burnAlert, exhaustionMinutes, "info", sloID),
			},
			// Import with invalid import ID
			{
				ResourceName:        "honeycombio_burn_alert.test",
				ImportStateIdPrefix: fmt.Sprintf("%v.", dataset),
				ImportState:         true,
				ExpectError:         regexp.MustCompile(`Error: Invalid Import ID`),
			},
		},
	})
}

func TestAcc_BurnAlertResource_validateDefault(t *testing.T) {
	dataset, sloID := burnAlertAccTestSetup(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
		CheckDestroy:             testAccEnsureBurnAlertDestroyed(t),
		Steps: []resource.TestStep{
			{
				Config:      testAccConfigBurnAlertDefault_basic(-1, dataset, sloID, "info"),
				ExpectError: regexp.MustCompile(`exhaustion_minutes value must be at least`),
			},
			{
				Config:      testAccConfigBurnAlertDefault_validateAttributesWhenAlertTypeIsExhaustionTime(dataset, sloID),
				ExpectError: regexp.MustCompile(`argument "exhaustion_minutes" is required`),
			},
			{
				Config:      testAccConfigBurnAlertDefault_validateAttributesWhenAlertTypeIsExhaustionTime(dataset, sloID),
				ExpectError: regexp.MustCompile(`"budget_rate_window_minutes": must not be configured when "alert_type"`),
			},
			{
				Config:      testAccConfigBurnAlertDefault_validateAttributesWhenAlertTypeIsExhaustionTime(dataset, sloID),
				ExpectError: regexp.MustCompile(`"budget_rate_decrease_percent": must not be configured when`),
			},
		},
	})
}

func TestAcc_BurnAlertResource_validateExhaustionTime(t *testing.T) {
	dataset, sloID := burnAlertAccTestSetup(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
		CheckDestroy:             testAccEnsureBurnAlertDestroyed(t),
		Steps: []resource.TestStep{
			{
				Config:      testAccConfigBurnAlertExhaustionTime_basic(-1, dataset, sloID, "info"),
				ExpectError: regexp.MustCompile(`exhaustion_minutes value must be at least`),
			},
			{
				Config:      testAccConfigBurnAlertExhaustionTime_validateAttributesWhenAlertTypeIsExhaustionTime(dataset, sloID),
				ExpectError: regexp.MustCompile(`argument "exhaustion_minutes" is required`),
			},
			{
				Config:      testAccConfigBurnAlertExhaustionTime_validateAttributesWhenAlertTypeIsExhaustionTime(dataset, sloID),
				ExpectError: regexp.MustCompile(`"budget_rate_window_minutes": must not be configured when "alert_type"`),
			},
			{
				Config:      testAccConfigBurnAlertExhaustionTime_validateAttributesWhenAlertTypeIsExhaustionTime(dataset, sloID),
				ExpectError: regexp.MustCompile(`"budget_rate_decrease_percent": must not be configured when`),
			},
		},
	})
}

func TestAcc_BurnAlertResource_validateBudgetRate(t *testing.T) {
	dataset, sloID := burnAlertAccTestSetup(t)

	budgetRateWindowMinutes := 60
	budgetRateDecreasePercent := float64(1)

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
		CheckDestroy:             testAccEnsureBurnAlertDestroyed(t),
		Steps: []resource.TestStep{
			{
				Config:      testAccConfigBurnAlertBudgetRate_basic(0, budgetRateDecreasePercent, dataset, sloID, "info"),
				ExpectError: regexp.MustCompile(`budget_rate_window_minutes value must be at least`),
			},
			{
				Config:      testAccConfigBurnAlertBudgetRate_basic(budgetRateWindowMinutes, float64(0), dataset, sloID, "info"),
				ExpectError: regexp.MustCompile(`budget_rate_decrease_percent value must be at least`),
			},
			{
				Config:      testAccConfigBurnAlertBudgetRate_basic(budgetRateWindowMinutes, float64(200), dataset, sloID, "info"),
				ExpectError: regexp.MustCompile(`budget_rate_decrease_percent value must be at most`),
			},
			{
				Config:      testAccConfigBurnAlertBudgetRate_basic(budgetRateWindowMinutes, float64(1.123456789), dataset, sloID, "info"),
				ExpectError: regexp.MustCompile(`budget_rate_decrease_percent precision for value must be at most`),
			},
			{
				Config:      testAccConfigBurnAlertBudgetRate_validateAttributesWhenAlertTypeIsBudgetRate(dataset, sloID),
				ExpectError: regexp.MustCompile(`argument "budget_rate_decrease_percent" is required`),
			},
			{
				Config:      testAccConfigBurnAlertBudgetRate_validateAttributesWhenAlertTypeIsBudgetRate(dataset, sloID),
				ExpectError: regexp.MustCompile(`argument "budget_rate_window_minutes" is required`),
			},
			{
				Config:      testAccConfigBurnAlertBudgetRate_validateAttributesWhenAlertTypeIsBudgetRate(dataset, sloID),
				ExpectError: regexp.MustCompile(`"exhaustion_minutes": must not be configured when "alert_type"`),
			},
		},
	})
}

func TestAcc_BurnAlertResource_validateUnknownOrVariableAttributesExhaustionTime(t *testing.T) {
	dataset, sloID := burnAlertAccTestSetup(t)
	burnAlert := &client.BurnAlert{}

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
		CheckDestroy:             testAccEnsureBurnAlertDestroyed(t),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigBurnAlertBudgetRate_validateUnknownOrVariableAttributesWhenAlertTypeIsExhaustionTime(dataset, sloID),
				Check:  testAccEnsureSuccessExhaustionTimeAlert(t, burnAlert, 90, "info", sloID),
			},
		},
	})
}

// Checks to ensure that if a Burn Alert was removed from Honeycomb outside of Terraform (UI or API)
// that it is detected and planned for recreation.
func TestAcc_BurnAlertResource_recreateOnNotFound(t *testing.T) {
	dataset, sloID := burnAlertAccTestSetup(t)
	burnAlert := &client.BurnAlert{}

	exhaustionMinutes := 240

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigBurnAlertExhaustionTime_basic(exhaustionMinutes, dataset, sloID, "info"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccEnsureSuccessExhaustionTimeAlert(t, burnAlert, exhaustionMinutes, "info", sloID),
					func(_ *terraform.State) error {
						// the final 'check' deletes the Burn Alert directly via the API leaving it behind in the state
						err := testAccClient(t).BurnAlerts.Delete(context.Background(), dataset, burnAlert.ID)
						if err != nil {
							return fmt.Errorf("failed to delete Burn Alert: %w", err)
						}
						return nil
					},
				),
				// ensure that the plan is non-empty after the deletion
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAcc_BurnAlertResource_HandlesRecipientChangedOutsideOfTerraform(t *testing.T) {
	c := testAccClient(t)
	ctx := context.Background()
	dataset, sloID := burnAlertAccTestSetup(t)

	// setup a slack recipient to be used in the burn alert, and modified outside of terraform
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
				Config: testAccConfigBurnAlertWithSlackRecipient(dataset, sloID, channel),
			},
			{
				PreConfig: func() {
					// update the channel name outside of Terraform
					channel += "-1"
					rcpt.Details.SlackChannel = channel
					_, err := c.Recipients.Update(ctx, rcpt)
					require.NoError(t, err, "failed to update test recipient")
				},
				Config: testAccConfigBurnAlertWithSlackRecipient(dataset, sloID, channel),
			},
		},
	})
}

// ensures no type error when using a dynamic recipient block
func TestAcc_BurnAlertResource_HandlesDynamicRecipientBlock(t *testing.T) {
	dataset, sloID := burnAlertAccTestSetup(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigBurnAlertWithDynamicRecipient(dataset, sloID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("honeycombio_burn_alert.test", "recipient.#", "2"),
				),
			},
		},
	})
}

func TestAcc_BurnAlertResource_HandlesDescriptionSetToEmptyString(t *testing.T) {
	ctx := context.Background()
	dataset, sloID := burnAlertAccTestSetup(t)
	burnAlert := &client.BurnAlert{}

	config := fmt.Sprintf(`
resource "honeycombio_burn_alert" "test" {
  exhaustion_minutes = 240

  dataset = "%s"
  slo_id  = "%s"

  recipient {
    type   = "email"
    target = "%s"
  }
}`, dataset, sloID, test.RandomEmail())

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccEnsureBurnAlertExists(t, "honeycombio_burn_alert.test", burnAlert),
					resource.TestCheckResourceAttr("honeycombio_burn_alert.test", "description", ""),
				),
			},
			{
				PreConfig: func() {
					// add a description to the burn alert outside of Terraform
					burnAlert.Description = "test description"
					_, err := testAccClient(t).BurnAlerts.Update(ctx, dataset, burnAlert)
					require.NoError(t, err, "failed to update burn alert")
				},
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccEnsureBurnAlertExists(t, "honeycombio_burn_alert.test", burnAlert),
					resource.TestCheckResourceAttr("honeycombio_burn_alert.test", "description", ""),
				),
			},
		},
	})
}

// Checks that the exhaustion time burn alert exists, has the correct attributes, and has the correct state
func testAccEnsureSuccessExhaustionTimeAlert(t *testing.T, burnAlert *client.BurnAlert, exhaustionMinutes int, pagerdutySeverity, sloID string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc(
		// Check that the burn alert exists
		testAccEnsureBurnAlertExists(t, "honeycombio_burn_alert.test", burnAlert),

		// Check that the burn alert has the correct attributes
		testAccEnsureAttributesCorrectExhaustionTime(burnAlert, exhaustionMinutes, sloID),

		// Check that the burn alert has the correct values in state
		resource.TestCheckResourceAttr("honeycombio_burn_alert.test", "slo_id", sloID),
		resource.TestCheckResourceAttr("honeycombio_burn_alert.test", "description", testBADescription),
		resource.TestCheckResourceAttr("honeycombio_burn_alert.test", "alert_type", "exhaustion_time"),
		resource.TestCheckResourceAttr("honeycombio_burn_alert.test", "exhaustion_minutes", fmt.Sprintf("%d", exhaustionMinutes)),
		resource.TestCheckResourceAttr("honeycombio_burn_alert.test", "recipient.#", "1"),
		resource.TestCheckResourceAttr("honeycombio_burn_alert.test", "recipient.0.notification_details.#", "1"),
		resource.TestCheckResourceAttr("honeycombio_burn_alert.test", "recipient.0.notification_details.0.pagerduty_severity", pagerdutySeverity),

		// Budget rate attributes should not be set
		resource.TestCheckNoResourceAttr("honeycombio_burn_alert.test", "budget_rate_window_minutes"),
		resource.TestCheckNoResourceAttr("honeycombio_burn_alert.test", "budget_rate_decrease_percent"),
	)
}

// Checks that the exhaustion time burn alert exists, has the correct attributes, and has the correct state
func testAccEnsureSuccessExhaustionTimeAlertWithWebhookRecip(t *testing.T, burnAlert *client.BurnAlert, exhaustionMinutes int, sloID, varValue string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc(
		// Check that the burn alert exists
		testAccEnsureBurnAlertExists(t, "honeycombio_burn_alert.test", burnAlert),

		// Check that the burn alert has the correct attributes
		testAccEnsureAttributesCorrectExhaustionTime(burnAlert, exhaustionMinutes, sloID),

		// Check that the burn alert has the correct values in state
		resource.TestCheckResourceAttr("honeycombio_burn_alert.test", "slo_id", sloID),
		resource.TestCheckResourceAttr("honeycombio_burn_alert.test", "description", testBADescription),
		resource.TestCheckResourceAttr("honeycombio_burn_alert.test", "alert_type", "exhaustion_time"),
		resource.TestCheckResourceAttr("honeycombio_burn_alert.test", "exhaustion_minutes", fmt.Sprintf("%d", exhaustionMinutes)),
		resource.TestCheckResourceAttr("honeycombio_burn_alert.test", "recipient.#", "1"),
		resource.TestCheckResourceAttr("honeycombio_burn_alert.test", "recipient.0.notification_details.#", "1"),
		resource.TestCheckResourceAttr("honeycombio_burn_alert.test", "recipient.0.notification_details.0.variable.0.name", "severity"),
		resource.TestCheckResourceAttr("honeycombio_burn_alert.test", "recipient.0.notification_details.0.variable.0.value", varValue),

		// Budget rate attributes should not be set
		resource.TestCheckNoResourceAttr("honeycombio_burn_alert.test", "budget_rate_window_minutes"),
		resource.TestCheckNoResourceAttr("honeycombio_burn_alert.test", "budget_rate_decrease_percent"),
	)
}

// Checks that the budget rate burn alert exists, has the correct attributes, and has the correct state
func testAccEnsureSuccessBudgetRateAlert(t *testing.T, burnAlert *client.BurnAlert, budgetRateWindowMinutes int, budgetRateDecreasePercent float64, pagerdutySeverity, sloID string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc(
		// Check that the burn alert exists
		testAccEnsureBurnAlertExists(t, "honeycombio_burn_alert.test", burnAlert),

		// Check that the burn alert has the correct attributes
		testAccEnsureAttributesCorrectBudgetRate(burnAlert, budgetRateWindowMinutes, budgetRateDecreasePercent, sloID),

		// Check that the burn alert has the correct values in state
		resource.TestCheckResourceAttr("honeycombio_burn_alert.test", "slo_id", sloID),
		resource.TestCheckResourceAttr("honeycombio_burn_alert.test", "description", testBADescription),
		resource.TestCheckResourceAttr("honeycombio_burn_alert.test", "alert_type", "budget_rate"),
		resource.TestCheckResourceAttr("honeycombio_burn_alert.test", "budget_rate_window_minutes", fmt.Sprintf("%d", budgetRateWindowMinutes)),
		resource.TestCheckResourceAttr("honeycombio_burn_alert.test", "budget_rate_decrease_percent", helper.FloatToPercentString(budgetRateDecreasePercent)),
		resource.TestCheckResourceAttr("honeycombio_burn_alert.test", "recipient.#", "1"),
		resource.TestCheckResourceAttr("honeycombio_burn_alert.test", "recipient.0.notification_details.#", "1"),
		resource.TestCheckResourceAttr("honeycombio_burn_alert.test", "recipient.0.notification_details.0.pagerduty_severity", pagerdutySeverity),
		// Exhaustion time attributes should not be set
		resource.TestCheckNoResourceAttr("honeycombio_burn_alert.test", "exhaustion_minutes"),
	)
}

// Checks that the budget rate burn alert exists, has the correct attributes, and has the correct state
func testAccEnsureSuccessBudgetRateAlertWithWebhookRecip(t *testing.T, burnAlert *client.BurnAlert, budgetRateWindowMinutes int, budgetRateDecreasePercent float64, sloID, varValue string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc(
		// Check that the burn alert exists
		testAccEnsureBurnAlertExists(t, "honeycombio_burn_alert.test", burnAlert),

		// Check that the burn alert has the correct attributes
		testAccEnsureAttributesCorrectBudgetRate(burnAlert, budgetRateWindowMinutes, budgetRateDecreasePercent, sloID),

		// Check that the burn alert has the correct values in state
		resource.TestCheckResourceAttr("honeycombio_burn_alert.test", "slo_id", sloID),
		resource.TestCheckResourceAttr("honeycombio_burn_alert.test", "description", testBADescription),
		resource.TestCheckResourceAttr("honeycombio_burn_alert.test", "alert_type", "budget_rate"),
		resource.TestCheckResourceAttr("honeycombio_burn_alert.test", "budget_rate_window_minutes", fmt.Sprintf("%d", budgetRateWindowMinutes)),
		resource.TestCheckResourceAttr("honeycombio_burn_alert.test", "budget_rate_decrease_percent", helper.FloatToPercentString(budgetRateDecreasePercent)),
		resource.TestCheckResourceAttr("honeycombio_burn_alert.test", "recipient.#", "1"),
		resource.TestCheckResourceAttr("honeycombio_burn_alert.test", "recipient.0.notification_details.#", "1"),
		resource.TestCheckResourceAttr("honeycombio_burn_alert.test", "recipient.0.notification_details.0.variable.#", "1"),
		resource.TestCheckResourceAttr("honeycombio_burn_alert.test", "recipient.0.notification_details.0.variable.0.name", "severity"),
		resource.TestCheckResourceAttr("honeycombio_burn_alert.test", "recipient.0.notification_details.0.variable.0.value", varValue),

		// Exhaustion time attributes should not be set
		resource.TestCheckNoResourceAttr("honeycombio_burn_alert.test", "exhaustion_minutes"),
	)
}

func testAccEnsureBurnAlertExists(t *testing.T, name string, burnAlert *client.BurnAlert) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resourceState, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("\"%s\" not found in state", name)
		}

		client := testAccClient(t)
		alert, err := client.BurnAlerts.Get(context.Background(), resourceState.Primary.Attributes["dataset"], resourceState.Primary.ID)
		if err != nil {
			return fmt.Errorf("failed to fetch created Burn Alert: %w", err)
		}

		*burnAlert = *alert

		return nil
	}
}

func testAccEnsureAttributesCorrectExhaustionTime(burnAlert *client.BurnAlert, exhaustionMinutes int, sloID string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if burnAlert.AlertType != "exhaustion_time" {
			return fmt.Errorf("incorrect alert_type: %s", burnAlert.AlertType)
		}

		if burnAlert.ExhaustionMinutes == nil {
			return fmt.Errorf("incorrect exhaustion_minutes: expected not to be nil")
		}
		if *burnAlert.ExhaustionMinutes != exhaustionMinutes {
			return fmt.Errorf("incorrect exhaustion_minutes: %d", *burnAlert.ExhaustionMinutes)
		}

		if burnAlert.SLO.ID != sloID {
			return fmt.Errorf("incorrect SLO ID: %s", burnAlert.SLO.ID)
		}

		// TODO: more in-depth checking of recipients

		return nil
	}
}

func testAccEnsureAttributesCorrectBudgetRate(burnAlert *client.BurnAlert, budgetRateWindowMinutes int, budgetRateDecreasePercent float64, sloID string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if burnAlert.AlertType != "budget_rate" {
			return fmt.Errorf("incorrect alert_type: %s", burnAlert.AlertType)
		}

		if burnAlert.BudgetRateWindowMinutes == nil {
			return fmt.Errorf("incorrect budget_rate_window_minutes: expected not to be nil")
		}
		if *burnAlert.BudgetRateWindowMinutes != budgetRateWindowMinutes {
			return fmt.Errorf("incorrect budget_rate_window_minutes: %d", *burnAlert.BudgetRateWindowMinutes)
		}

		if burnAlert.BudgetRateDecreaseThresholdPerMillion == nil {
			return fmt.Errorf("incorrect budget_rate_decrease_percent: expected not to be nil")
		}
		// Must convert from PPM back to float to compare with config
		budgetRateDecreaseThresholdPerMillionAsPercent := helper.PPMToFloat(*burnAlert.BudgetRateDecreaseThresholdPerMillion)
		if budgetRateDecreaseThresholdPerMillionAsPercent != budgetRateDecreasePercent {
			return fmt.Errorf("incorrect budget_rate_decrease_percent: %f", budgetRateDecreaseThresholdPerMillionAsPercent)
		}

		if burnAlert.SLO.ID != sloID {
			return fmt.Errorf("incorrect SLO ID: %s", burnAlert.SLO.ID)
		}

		// TODO: more in-depth checking of recipients

		return nil
	}
}

func testAccEnsureBurnAlertDestroyed(t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, resourceState := range s.RootModule().Resources {
			if resourceState.Type != "honeycomb_burn_alert" {
				continue
			}

			if resourceState.Primary.ID == "" {
				return fmt.Errorf("no ID set for burn alert")
			}

			client := testAccClient(t)
			_, err := client.BurnAlerts.Get(context.Background(), resourceState.Primary.Attributes["dataset"], resourceState.Primary.ID)
			if err == nil {
				return fmt.Errorf("burn alert %s was not deleted on destroy", resourceState.Primary.ID)
			}
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
	require.NoError(t, err)

	slo, err := c.SLOs.Create(ctx, dataset, &client.SLO{
		Name:             acctest.RandString(8) + " SLO",
		TimePeriodDays:   14,
		TargetPerMillion: 995000,
		SLI:              client.SLIRef{Alias: sli.Alias},
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		// remove SLO, SLI DC at end of test run
		c.SLOs.Delete(ctx, dataset, slo.ID)
		c.DerivedColumns.Delete(ctx, dataset, sli.ID)
	})

	return dataset, slo.ID
}

func testAccConfigBurnAlert_withoutDescription(exhaustionMinutes int, dataset, sloID, pdseverity string) string {
	return fmt.Sprintf(`
resource "honeycombio_pagerduty_recipient" "test" {
  integration_key  = "08b9d4cacd68933151a1ef1028b67da2"
  integration_name = "test.pd-basic"
}

resource "honeycombio_burn_alert" "test" {
  exhaustion_minutes = %[1]d

  dataset            = "%[2]s"
  slo_id             = "%[3]s"
  recipient {
    id = honeycombio_pagerduty_recipient.test.id

    notification_details {
      pagerduty_severity = "%[4]s"
    }
  }
}`, exhaustionMinutes, dataset, sloID, pdseverity)
}

func testAccConfigBurnAlertDefault_basic(exhaustionMinutes int, dataset, sloID, pdseverity string) string {
	tmplBody := `<<EOT
		{
			"name": " {{ .Name }}",
			"id": " {{ .ID }}",
			"description": " {{ .Description }}",
		}
		EOT`
	return fmt.Sprintf(`
resource "honeycombio_pagerduty_recipient" "test" {
  integration_key  = "08b9d4cacd68933151a1ef1028b67da2"
  integration_name = "test.pd-basic"
}

resource "honeycombio_burn_alert" "test" {
  exhaustion_minutes = %[1]d

  dataset            = "%[2]s"
  slo_id             = "%[3]s"
  description        = "%[5]s"
  
  recipient {
    id = honeycombio_pagerduty_recipient.test.id

    notification_details {
      pagerduty_severity = "%[4]s"
    }
  }
}`, exhaustionMinutes, dataset, sloID, pdseverity, testBADescription, tmplBody)
}

func testAccConfigBurnAlertExhaustionTime_basicWebhookRecipient(exhaustionMinutes int, dataset, sloID, variableValue string) string {
	tmplBody := `<<EOT
		{
			"name": " {{ .Name }}",
			"id": " {{ .ID }}",
			"description": " {{ .Description }}",
		}
		EOT`
	return fmt.Sprintf(`
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
	  type   = "exhaustion_time"
      body = %[5]s
    }
}

resource "honeycombio_burn_alert" "test" {
  exhaustion_minutes = %[1]d

  dataset            = "%[2]s"
  slo_id             = "%[3]s"
  description        = "%[4]s"

  recipient {
	id = honeycombio_webhook_recipient.test.id
	
	notification_details {	
		variable {
			name = "severity"
			value = "%[6]s"
		}
	}
  }
}`, exhaustionMinutes, dataset, sloID, testBADescription, tmplBody, variableValue)
}

func testAccConfigBurnAlertDefault_validateAttributesWhenAlertTypeIsExhaustionTime(dataset, sloID string) string {
	return fmt.Sprintf(`
resource "honeycombio_burn_alert" "test" {
  budget_rate_window_minutes   = 60
  budget_rate_decrease_percent = 1

  dataset = "%[1]s"
  slo_id  = "%[2]s"

  recipient {
    type   = "email"
    target = "%s[3]s"
  }
}`, dataset, sloID, test.RandomEmail())
}

func testAccConfigBurnAlertExhaustionTime_basic(exhaustionMinutes int, dataset, sloID, pdseverity string) string {
	return fmt.Sprintf(`
resource "honeycombio_pagerduty_recipient" "test" {
  integration_key  = "08b9d4cacd68933151a1ef1028b67da2"
  integration_name = "test.pd-basic"
}

resource "honeycombio_burn_alert" "test" {
  alert_type         = "exhaustion_time"
  description        = "%[5]s"
  exhaustion_minutes = %[1]d

  dataset = "%[2]s"
  slo_id  = "%[3]s"

  recipient {
    id = honeycombio_pagerduty_recipient.test.id

    notification_details {
      pagerduty_severity = "%[4]s"
    }
  }
}`, exhaustionMinutes, dataset, sloID, pdseverity, testBADescription)
}

func testAccConfigBurnAlertExhaustionTime_validateAttributesWhenAlertTypeIsExhaustionTime(dataset, sloID string) string {
	return fmt.Sprintf(`
resource "honeycombio_burn_alert" "test" {
  alert_type                   = "exhaustion_time"
  budget_rate_window_minutes   = 60
  budget_rate_decrease_percent = 1

  dataset = "%[1]s"
  slo_id  = "%[2]s"

  recipient {
    type   = "email"
    target = "%[3]s"
  }
}`, dataset, sloID, test.RandomEmail())
}

func testAccConfigBurnAlertBudgetRate_basic(budgetRateWindowMinutes int, budgetRateDecreasePercent float64, dataset, sloID, pdseverity string) string {
	return fmt.Sprintf(`
resource "honeycombio_pagerduty_recipient" "test" {
  integration_key  = "08b9d4cacd68933151a1ef1028b67da2"
  integration_name = "test.pd-basic"
}

resource "honeycombio_burn_alert" "test" {
  alert_type                   = "budget_rate"
  description                  = "%[6]s"
  budget_rate_window_minutes   = %[1]d
  budget_rate_decrease_percent = %[2]s

  dataset = "%[3]s"
  slo_id  = "%[4]s"

  recipient {
    id = honeycombio_pagerduty_recipient.test.id

    notification_details {
      pagerduty_severity = "%[5]s"
    }
  }
}`, budgetRateWindowMinutes, helper.FloatToPercentString(budgetRateDecreasePercent), dataset, sloID, pdseverity, testBADescription)
}

func testAccConfigBurnAlertBudgetRate_basicWebhookRecipient(budgetRateWindowMinutes int, budgetRateDecreasePercent float64, dataset, sloID, variableValue string) string {
	tmplBody := `<<EOT
		{
			"name": " {{ .Name }}",
			"id": " {{ .ID }}",
			"description": " {{ .Description }}",
		}
		EOT`
	return fmt.Sprintf(`
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
	  type   = "budget_rate"
      body = %[6]s
    }
}

resource "honeycombio_burn_alert" "test" {
  alert_type                   = "budget_rate"
  description                  = "%[5]s"
  budget_rate_window_minutes   = %[1]d
  budget_rate_decrease_percent = %[2]s

  dataset = "%[3]s"
  slo_id  = "%[4]s"

  recipient {
	id = honeycombio_webhook_recipient.test.id
	
	notification_details {	
		variable {
			name = "severity"
			value = "%[7]s"
		}
	}
  }
}`, budgetRateWindowMinutes, helper.FloatToPercentString(budgetRateDecreasePercent), dataset, sloID, testBADescription, tmplBody, variableValue)
}

func testAccConfigBurnAlertBudgetRate_trailingZeros(dataset, sloID string) string {
	return fmt.Sprintf(`
resource "honeycombio_pagerduty_recipient" "test" {
  integration_key  = "08b9d4cacd68933151a1ef1028b67da2"
  integration_name = "test.pd-basic"
}

resource "honeycombio_burn_alert" "test" {
  alert_type                   = "budget_rate"
  description				   = "%[3]s"
  budget_rate_window_minutes   = 60
  budget_rate_decrease_percent = 5.00000

  dataset = "%[1]s"
  slo_id  = "%[2]s"

  recipient {
    id = honeycombio_pagerduty_recipient.test.id

    notification_details {
      pagerduty_severity = "info"
    }
  }
}`, dataset, sloID, testBADescription)
}

func testAccConfigBurnAlertBudgetRate_validateAttributesWhenAlertTypeIsBudgetRate(dataset, sloID string) string {
	return fmt.Sprintf(`
resource "honeycombio_burn_alert" "test" {
  alert_type         = "budget_rate"
  exhaustion_minutes = 60

  dataset = "%[1]s"
  slo_id  = "%[2]s"

  recipient {
    type   = "email"
    target = "%[3]s"
  }
}`, dataset, sloID, test.RandomEmail())
}

func testAccConfigBurnAlertBudgetRate_validateUnknownOrVariableAttributesWhenAlertTypeIsExhaustionTime(dataset, sloID string) string {
	return fmt.Sprintf(`
variable "exhaustion_minutes" {
  type    = number
  default = 90
}

resource "honeycombio_pagerduty_recipient" "test" {
  integration_key  = "08b9d4cacd68933151a1ef1028b67da2"
  integration_name = "test.pd-basic"
}

resource "honeycombio_burn_alert" "test" {
  exhaustion_minutes = var.exhaustion_minutes

  dataset = "%[1]s"
  slo_id  = "%[2]s"
  description = "%[3]s"

  recipient {
    id = honeycombio_pagerduty_recipient.test.id

    notification_details {
      pagerduty_severity = "info"
    }
  }
}`, dataset, sloID, testBADescription)
}

func testAccConfigBurnAlertWithSlackRecipient(dataset, sloID, channel string) string {
	return fmt.Sprintf(`
resource "honeycombio_burn_alert" "test" {
  exhaustion_minutes = 60

  dataset = "%[1]s"
  slo_id  = "%[2]s"
  description = "%[4]s"

  recipient {
    type   = "slack"
    target = "%[3]s"
  }
}`, dataset, sloID, channel, testBADescription)
}

func testAccConfigBurnAlertWithDynamicRecipient(dataset, sloID string) string {
	return fmt.Sprintf(`
variable "recipients" {
  type = list(object({
    type   = string
    target = string
  }))

  default = [
    {
      "type": "email",
      "target": "%[3]s"
    },
    {
      "type": "email",
      "target": "%[4]s"
    }
  ]
}

resource "honeycombio_burn_alert" "test" {
  exhaustion_minutes = 60

  dataset = "%[1]s"
  slo_id  = "%[2]s"

  dynamic "recipient" {
    for_each = var.recipients

    content {
      type   = recipient.value.type
      target = recipient.value.target
    }
  }
}`, dataset, sloID, test.RandomEmail(), test.RandomEmail())
}
