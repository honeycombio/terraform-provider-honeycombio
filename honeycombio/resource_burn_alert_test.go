package honeycombio

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccHoneycombioBurnAlert_basic(t *testing.T) {
	ctx := context.Background()
	c := testAccClient(t)
	dataset := testAccDataset()

	sli, err := c.DerivedColumns.Create(ctx, dataset, &honeycombio.DerivedColumn{
		Alias:      "sli.acc_ba_test",
		Expression: "LT($duration_ms, 1000)",
	})
	if err != nil {
		t.Error(err)
	}
	slo, err := c.SLOs.Create(ctx, dataset, &honeycombio.SLO{
		Name:             "BA TestAcc SLO",
		TimePeriodDays:   14,
		TargetPerMillion: 995000,
		SLI:              honeycombio.SLIRef{Alias: sli.Alias},
	})
	require.NoError(t, err)
	// remove SLO, SLI DC at end of test run
	t.Cleanup(func() {
		c.SLOs.Delete(ctx, dataset, slo.ID)
		c.DerivedColumns.Delete(ctx, dataset, sli.ID)
	})

	resource.Test(t, resource.TestCase{
		PreCheck:          testAccPreCheck(t),
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "honeycombio_burn_alert" "test" {
  dataset            = "%s"
  slo_id             = "%s"
  exhaustion_minutes = 240 # 4 hours

  recipient {
    type   = "slack"
    target = "#test2"
  }

}
`, dataset, slo.ID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBurnAlertExists(t, dataset, "honeycombio_burn_alert.test"),
				),
			},
		},
	})
}

// add a recipient by ID to verify the diff is stable
func TestAccHoneycombioBurnAlert_RecipientById(t *testing.T) {
	ctx := context.Background()
	c := testAccClient(t)
	dataset := testAccDataset()

	sli, err := c.DerivedColumns.Create(ctx, dataset, &honeycombio.DerivedColumn{
		Alias:      "sli.acc_ba_test",
		Expression: "LT($duration_ms, 1000)",
	})
	if err != nil {
		t.Error(err)
	}
	slo, err := c.SLOs.Create(ctx, dataset, &honeycombio.SLO{
		Name:             "BA TestAcc SLO",
		TimePeriodDays:   14,
		TargetPerMillion: 995000,
		SLI:              honeycombio.SLIRef{Alias: sli.Alias},
	})
	require.NoError(t, err)
	// remove SLO, SLI DC at end of test run
	t.Cleanup(func() {
		c.SLOs.Delete(ctx, dataset, slo.ID)
		c.DerivedColumns.Delete(ctx, dataset, sli.ID)
	})

	trigger, deleteFn := createTriggerWithRecipient(t, dataset, honeycombio.TriggerRecipient{
		Type:   honeycombio.TriggerRecipientTypeEmail,
		Target: "acctest@example.com",
	})
	defer deleteFn()

	resource.Test(t, resource.TestCase{
		PreCheck:          testAccPreCheck(t),
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "honeycombio_burn_alert" "test" {
				  dataset            = "%s"
				  slo_id             = "%s"
				  exhaustion_minutes = 240 # 4 hours
				
				  recipient {
					id   = "%s"
				  }
				
				}
				`, dataset, slo.ID, trigger.Recipients[0].ID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBurnAlertExists(t, dataset, "honeycombio_burn_alert.test"),
				),
			},
		},
	})
}

func testAccCheckBurnAlertExists(t *testing.T, dataset string, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resourceState, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		client := testAccClient(t)
		createdBA, err := client.BurnAlerts.Get(context.Background(), dataset, resourceState.Primary.ID)
		if err != nil {
			return fmt.Errorf("could not find created BurnAlert: %w", err)
		}

		assert.Equal(t, resourceState.Primary.ID, createdBA.ID)
		assert.Equal(t, resourceState.Primary.Attributes["slo_id"], createdBA.SLO.ID)
		assert.Equal(t, resourceState.Primary.Attributes["exhaustion_minutes"], fmt.Sprintf("%v", createdBA.ExhaustionMinutes))
		assert.NotNil(t, resourceState.Primary.Attributes["recipient"])

		return nil
	}
}
