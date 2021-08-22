package honeycombio

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
)

func TestAccHoneycombioDataset_basic(t *testing.T) {
	testDataset := testAccDataset()

	resource.Test(t, resource.TestCase{
		PreCheck:          testAccPreCheck(t),
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasetConfig(testDataset),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatasetExists(t, "honeycombio_dataset.test", testDataset),
					resource.TestCheckResourceAttr("honeycombio_dataset.test", "name", testDataset),
					resource.TestCheckResourceAttr("honeycombio_dataset.test", "slug", urlEncodeDataset(testDataset)),
				),
			},
		},
	})
}

func testAccDatasetConfig(dataset string) string {
	return fmt.Sprintf(`
resource "honeycombio_dataset" "test" {
  name = "%s"
}`, dataset)
}

// testAccCheckDatasetExists queries the API to verify the Dataset exists and
// matches with the Terraform state.
func testAccCheckDatasetExists(t *testing.T, name, testDataset string) resource.TestCheckFunc {
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

		assert.Equal(t, testDataset, d.Name)
		assert.Equal(t, urlEncodeDataset(testDataset), d.Slug)

		return nil
	}
}

// urlEncodeDataset sanitizes the dataset name for when it is used as part of
// the URL. This matches with how Honeycomb creates the slug version of a name.
func urlEncodeDataset(dataset string) string {
	return strings.Replace(dataset, "/", "-", -1)
}
