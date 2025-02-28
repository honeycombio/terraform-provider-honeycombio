package honeycombio

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/require"

	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/test"
)

func TestAccHoneycombioSLO_basic(t *testing.T) {
	dataset, sliAlias := sloAccTestSetup(t)
	slo := &honeycombio.SLO{}

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigSLO_basic(dataset, sliAlias),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSLOExists(t, "honeycombio_slo.test", slo),
					resource.TestCheckResourceAttr("honeycombio_slo.test", "name", "TestAcc SLO"),
					resource.TestCheckResourceAttr("honeycombio_slo.test", "description", "integration test SLO"),
					resource.TestCheckResourceAttr("honeycombio_slo.test", "sli", sliAlias),
					resource.TestCheckResourceAttr("honeycombio_slo.test", "target_percentage", "99.95"),
					resource.TestCheckResourceAttr("honeycombio_slo.test", "time_period", "30"),
				),
			},
		},
	})
}

// Checks to ensure that if an SLO was removed from Honeycomb outside of Terraform (UI or API)
// that it is detected and planned for recreation.
func TestAccHoneycombioSLO_RecreateOnNotFound(t *testing.T) {
	dataset, sliAlias := sloAccTestSetup(t)
	slo := &honeycombio.SLO{}

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigSLO_basic(dataset, sliAlias),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSLOExists(t, "honeycombio_slo.test", slo),
					func(_ *terraform.State) error {
						// the final 'check' deletes the SLO directly via the API leaving it behind in the state
						err := testAccClient(t).SLOs.Delete(context.Background(), dataset, slo.ID)
						if err != nil {
							return fmt.Errorf("failed to delete SLO: %w", err)
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

func testAccConfigSLO_basic(dataset, sliAlias string) string {
	return fmt.Sprintf(`
	resource "honeycombio_slo" "test" {
		name              = "TestAcc SLO"
		description       = "integration test SLO"
		dataset           = "%s"
		sli               = "%s"
		target_percentage = 99.95
		time_period       = 30
	}
	`, dataset, sliAlias)
}

func testAccCheckSLOExists(t *testing.T, name string, slo *honeycombio.SLO) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resourceState, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("\"%s\" not found in state", name)
		}

		client := testAccClient(t)
		rslo, err := client.SLOs.Get(context.Background(), resourceState.Primary.Attributes["dataset"], resourceState.Primary.ID)
		if err != nil {
			return fmt.Errorf("failed to fetch created SLO: %w", err)
		}

		*slo = *rslo

		return nil
	}
}

func sloAccTestSetup(t *testing.T) (string, string) {
	t.Helper()

	ctx := context.Background()
	c := testAccClient(t)
	dataset := testAccDataset()

	sli, err := c.DerivedColumns.Create(ctx, dataset, &honeycombio.DerivedColumn{
		Alias:      test.RandomStringWithPrefix("test.", 8),
		Expression: "BOOL(1)",
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		// remove SLI DC at end of test run
		c.DerivedColumns.Delete(ctx, dataset, sli.ID)
	})

	return dataset, sli.Alias
}
