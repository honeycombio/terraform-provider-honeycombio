package honeycombio

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccHoneycombioDataset_basic(t *testing.T) {
	testDatasetName := "testacc-test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasetConfig(testDatasetName, "a nice description", 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatasetExists(t, "honeycombio_dataset.test"),
					resource.TestCheckResourceAttr("honeycombio_dataset.test", "name", testDatasetName),
					resource.TestCheckResourceAttr("honeycombio_dataset.test", "description", "a nice description"),
					resource.TestCheckResourceAttr("honeycombio_dataset.test", "expand_json_depth", "3"),
				),
			},
			{
				ResourceName:      "honeycombio_dataset.test",
				ImportStateId:     testDatasetName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccDatasetConfig(name, description string, depth int) string {
	return fmt.Sprintf(`
resource "honeycombio_dataset" "test" {
  name              = "%s"
  description       = "%s"
  expand_json_depth = %d
}`, name, description, depth)
}

// testAccCheckDatasetExists queries the API to verify the Dataset exists
func testAccCheckDatasetExists(t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resourceState, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		c := testAccClient(t)
		_, err := c.Datasets.Get(context.Background(), resourceState.Primary.ID)
		if err != nil {
			return fmt.Errorf("could not retrieve dataset: %w", err)
		}

		return nil
	}
}
