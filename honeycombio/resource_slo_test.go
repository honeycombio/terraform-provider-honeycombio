package honeycombio

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/stretchr/testify/require"

	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
)

func TestAccHoneycombioSLO_basic(t *testing.T) {
	ctx := context.Background()
	c := testAccClient(t)
	dataset := testAccDataset()

	sli, err := c.DerivedColumns.Create(ctx, dataset, &honeycombio.DerivedColumn{
		Alias:      "sli.acc_slo_test",
		Expression: "LT($duration_ms, 1000)",
	})
	require.NoError(t, err)
	//nolint:errcheck
	t.Cleanup(func() {
		// remove SLI DC at end of test run
		c.DerivedColumns.Delete(ctx, dataset, sli.ID)
	})

	resource.Test(t, resource.TestCase{
		PreCheck:          testAccPreCheck(t),
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "honeycombio_slo" "test" {
  name              = "TestAcc SLO"
  description       = "integration test SLO"
  dataset           = "%s"
  sli               = "%s"
  target_percentage = 99.95
  time_period       = 30
}
`, dataset, sli.Alias),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSLOExists(t, dataset, "honeycombio_slo.test"),
					resource.TestCheckResourceAttr("honeycombio_slo.test", "name", "TestAcc SLO"),
					resource.TestCheckResourceAttr("honeycombio_slo.test", "description", "integration test SLO"),
					resource.TestCheckResourceAttr("honeycombio_slo.test", "sli", sli.Alias),
					resource.TestCheckResourceAttr("honeycombio_slo.test", "target_percentage", "99.95"),
					resource.TestCheckResourceAttr("honeycombio_slo.test", "time_period", "30"),
				),
			},
		},
	})
}

func testAccCheckSLOExists(t *testing.T, dataset string, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resourceState, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		client := testAccClient(t)
		_, err := client.SLOs.Get(context.Background(), dataset, resourceState.Primary.ID)
		if err != nil {
			return fmt.Errorf("could not find created SLO: %w", err)
		}

		return nil
	}
}
