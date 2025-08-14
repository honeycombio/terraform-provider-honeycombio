package honeycombio

import (
	"fmt"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/test"
)

func TestAccHoneycombioDatasetDefinition_basic(t *testing.T) {
	dataset := testAccDataset()
	col1Name := test.RandomStringWithPrefix("test.", 8)
	col2Name := test.RandomStringWithPrefix("test.", 8)
	col3Name := test.RandomStringWithPrefix("test.", 8)

	config := fmt.Sprintf(`
resource "honeycombio_column" "column_1" {
  dataset = "%[1]s"
  name    = "%[2]s"
}

resource "honeycombio_column" "column_2" {
  dataset = "%[1]s"
  name    = "%[3]s"
}

resource "honeycombio_column" "column_3" {
  dataset = "%[1]s"
  name    = "%[4]s"
}

resource "honeycombio_derived_column" "log10_duration" {
  dataset = "%[1]s"

  alias      = "log10_duration"
  expression = "LOG10($duration_ms)"
}

resource "honeycombio_dataset_definition" "name" {
  dataset = "%[1]s"

  name   = "name"
  column = honeycombio_column.column_1.name
}

resource "honeycombio_dataset_definition" "duration_ms" {
  dataset = "%[1]s"

  name   = "duration_ms"
  column = honeycombio_derived_column.log10_duration.alias
}

resource "honeycombio_dataset_definition" "route" {
  dataset = "%[1]s"

  name   = "route"
  column = honeycombio_column.column_2.name
}

resource "honeycombio_dataset_definition" "error" {
  dataset = "%[1]s"

  name   = "error"
  column = honeycombio_column.column_3.name
}
`, dataset, col1Name, col2Name, col3Name)

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
		Steps: []resource.TestStep{
			{
				Config: config,
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
			{
				// remove the 'error' definition and ensure reading the definitions still works
				PreConfig: func() {
					ctx := t.Context()
					client := testAccClient(t)
					client.DatasetDefinitions.Update(ctx, dataset, &honeycombio.DatasetDefinition{
						Error: &honeycombio.DefinitionColumn{Name: ""},
					})
				},
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("honeycombio_dataset_definition.error", "name", "error"),
					resource.TestCheckResourceAttr("honeycombio_dataset_definition.error", "column", col3Name),
					resource.TestCheckResourceAttr("honeycombio_dataset_definition.error", "column_type", "column"),
				),
			},
		},
		CheckDestroy: func(s *terraform.State) error {
			// ensure that after destroying ('deleting') the above definitions
			// they have been reset to their default values
			client := testAccClient(t)
			dd, err := client.DatasetDefinitions.Get(t.Context(), dataset)
			if err != nil {
				return fmt.Errorf("could not lookup dataset definitions: %w", err)
			}
			for _, defn := range []string{"name", "duration_ms", "route", "error"} {
				column := extractDatasetDefinitionColumnByName(dd, defn)
				if column == nil {
					continue // definition was removed and does not have default
				}
				if slices.Contains(honeycombio.DatasetDefinitionDefaults()[defn], column.Name) {
					continue // definition was removed and reset to default
				}
				return fmt.Errorf("definition %q was not reset to default: %v", defn, column)
			}
			return nil
		},
	})
}
