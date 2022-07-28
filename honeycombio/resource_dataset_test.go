package honeycombio

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/stretchr/testify/assert"
)

func TestAccHoneycombioDataset_basic(t *testing.T) {

	createArgs := honeycombio.DatasetCreateArgs{
		Description:     "buzzing with data",
		ExpandJSONDepth: 3,
	}
	testDataset := testAccDatasetWithArgs(createArgs)

	resource.Test(t, resource.TestCase{
		PreCheck:          testAccPreCheck(t),
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasetConfig(testDataset.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatasetExists(t, "honeycombio_dataset.test", testDataset.Name),
					resource.TestCheckResourceAttr("honeycombio_dataset.test", "name", testDataset.Name),
					resource.TestCheckResourceAttr("honeycombio_dataset.test", "description", *testDataset.Description),
					resource.TestCheckResourceAttr("honeycombio_dataset.test", "slug", urlEncodeDataset(testDataset.Slug)),
					resource.TestCheckResourceAttr("honeycombio_dataset.test", "expand_json_depth", fmt.Sprintf("%d", *testDataset.ExpandJSONDepth)),
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
		assert.Equal(t, testDataset, d.Description)
		assert.Equal(t, urlEncodeDataset(testDataset), d.Slug)
		assert.Equal(t, testDataset, d.ExpandJSONDepth)

		return nil
	}
}

// urlEncodeDataset sanitizes the dataset name for when it is used as part of
// the URL. This matches with how Honeycomb creates the slug version of a name.
func urlEncodeDataset(dataset string) string {
	return strings.Replace(dataset, "/", "-", -1)
}
