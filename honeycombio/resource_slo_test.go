package honeycombio

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/stretchr/testify/assert"
)

func TestAccHoneycombioSLO_basic(t *testing.T) {
	ctx := context.Background()
	c := testAccClient(t)
	dataset := testAccDataset()

	sli, err := c.DerivedColumns.Create(ctx, dataset, &honeycombio.DerivedColumn{
		Alias:      "sli.acc_slo_test",
		Expression: "LT($duration_ms, 1000)",
	})
	if err != nil {
		t.Error(err)
	}
	// remove SLI DC at end of test run
	t.Cleanup(func() {
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
					testAccCheckSLOExists(t, dataset, "honeycombio_slo.test", "TestAcc SLO"),
				),
			},
		},
	})
}

func testAccCheckSLOExists(t *testing.T, dataset string, resourceName string, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resourceState, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		client := testAccClient(t)
		createdSLO, err := client.SLOs.Get(context.Background(), dataset, resourceState.Primary.ID)
		if err != nil {
			return fmt.Errorf("could not find created SLO: %w", err)
		}

		assert.Equal(t, resourceState.Primary.ID, createdSLO.ID)
		assert.Equal(t, name, createdSLO.Name)
		assert.Equal(t, resourceState.Primary.Attributes["description"], createdSLO.Description)
		assert.Equal(t, resourceState.Primary.Attributes["sli"], createdSLO.SLI.Alias)
		assert.Equal(t, resourceState.Primary.Attributes["target_percentage"], fmt.Sprintf("%v", tpmToFloat(createdSLO.TargetPerMillion)))
		assert.Equal(t, resourceState.Primary.Attributes["time_period"], fmt.Sprintf("%v", createdSLO.TimePeriodDays))

		return nil
	}
}
