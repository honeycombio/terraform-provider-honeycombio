package honeycombio

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slices"
)

func TestAccHoneycombioDatasetDefinition_basic(t *testing.T) {
	dataset := testAccDataset()

	resource.Test(t, resource.TestCase{
		PreCheck:          testAccPreCheck(t),
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "honeycombio_dataset_definition" "test" {
  dataset = "%s"

  name  = "%s"
  column = "%s"
}`, dataset, "name", "app.tenant"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("honeycombio_dataset_definition.test", "name", "name"),
					resource.TestCheckResourceAttr("honeycombio_dataset_definition.test", "column", "app.tenant"),
				),
			},
		},
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckDatasetDefinitionResetToDefault(t, dataset, "name"),
		),
	})
}

func testAccCheckDatasetDefinitionResetToDefault(t *testing.T, dataset string, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccClient(t)
		dd, err := client.DatasetDefinitions.Get(context.Background(), dataset)
		if err != nil {
			return fmt.Errorf("could not lookup dataset definitions: %w", err)
		}

		column := extractDatasetDefinitionByName(name, dd)
		assert.True(t, slices.Contains(honeycombio.DatasetDefinitionColumnDefaults()[name], column))

		return nil
	}
}
