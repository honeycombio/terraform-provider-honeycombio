package honeycombio

import (
	"context"
	"fmt"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/test"

	"github.com/stretchr/testify/assert"
)

func TestAccHoneycombioDatasetDefinition_basic(t *testing.T) {
	dataset := testAccDataset()
	col1Name := test.RandomStringWithPrefix("test.", 8)
	col2Name := test.RandomStringWithPrefix("test.", 8)

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "honeycombio_column" "column_1" {
  dataset = "%[1]s"
  name    = "%[2]s"
}

resource "honeycombio_column" "column_2" {
  dataset = "%[1]s"
  name    = "%[3]s"
}

resource "honeycombio_derived_column" "log10_duration" {
  dataset = "%[1]s"

  alias      = "log10_duration"
  expression = "LOG10($duration_ms)"
}

resource "honeycombio_dataset_definition" "name" {
  dataset = "%[1]s"

  name   = "name"
  column = honeycombio_column.column_1.key_name
}

resource "honeycombio_dataset_definition" "duration_ms" {
  dataset = "%[1]s"

  name   = "duration_ms"
  column = honeycombio_derived_column.log10_duration.alias
}

resource "honeycombio_dataset_definition" "route" {
  dataset = "%[1]s"

  name   = "route"
  column = honeycombio_column.column_2.key_name
}
`, dataset, col1Name, col2Name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("honeycombio_dataset_definition.name", "name", "name"),
					resource.TestCheckResourceAttr("honeycombio_dataset_definition.name", "column", col1Name),
					resource.TestCheckResourceAttr("honeycombio_dataset_definition.name", "column_type", "column"),
					resource.TestCheckResourceAttr("honeycombio_dataset_definition.duration_ms", "name", "duration_ms"),
					resource.TestCheckResourceAttr("honeycombio_dataset_definition.duration_ms", "column", "log10_duration"),
					resource.TestCheckResourceAttr("honeycombio_dataset_definition.duration_ms", "column_type", "derived_column"),
					resource.TestCheckResourceAttr("honeycombio_dataset_definition.route", "name", "route"),
					resource.TestCheckResourceAttr("honeycombio_dataset_definition.route", "column", col2Name),
					resource.TestCheckResourceAttr("honeycombio_dataset_definition.route", "column_type", "column"),
				),
			},
		},
		CheckDestroy: resource.ComposeTestCheckFunc(
			// ensure that after destroying ('deleting') the above definitions
			// they have been reset to their default values
			testAccCheckDatasetDefinitionResetToDefault(t, dataset, "name"),
			testAccCheckDatasetDefinitionResetToDefault(t, dataset, "duration_ms"),
			testAccCheckDatasetDefinitionResetToDefault(t, dataset, "route"),
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

		// defaults would be nil or one of the values in `DatasetDefinitionColumnDefaults`
		column := extractDatasetDefinitionColumnByName(dd, name)
		if column != nil {
			assert.True(t, slices.Contains(honeycombio.DatasetDefinitionDefaults()[name], column.Name))
		}

		return nil
	}
}
