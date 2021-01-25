package honeycombio

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
)

func TestAccHoneycombioDataset_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          testAccPreCheck(t),
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasetConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatasetExists(t, "honeycombio_dataset.test"),
					resource.TestCheckResourceAttr("honeycombio_dataset.test", "name", "kvrhdn/terraform-provider-honeycombio"),
					resource.TestCheckResourceAttr("honeycombio_dataset.test", "slug", "kvrhdn-terraform-provider-honeycombio"),
				),
			},
		},
	})
}

const testAccDatasetConfig = `
resource "honeycombio_dataset" "test" {
  name = "kvrhdn/terraform-provider-honeycombio"
}`

// testAccCheckDatasetExists queries the API to verify the Dataset exists and
// matches with the Terraform state.
func testAccCheckDatasetExists(t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resourceState, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		c := testAccClient(t)
		d, err := c.Datasets.Get(context.Background(), resourceState.Primary.ID)
		if err != nil {
			return fmt.Errorf("could not retrieve dataset: %w", err)
		}

		assert.Equal(t, "kvrhdn/terraform-provider-honeycombio", d.Name)
		assert.Equal(t, "kvrhdn-terraform-provider-honeycombio", d.Slug)

		return nil
	}
}
