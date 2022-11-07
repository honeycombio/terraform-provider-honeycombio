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

func TestAccHoneycombioDatasetDefinition_basic(t *testing.T) {
	// set multiple definitions in a single HCL block
	ctx := context.Background()
	c := testAccClient(t)
	dataset := testAccDataset()

	col, err := c.Columns.Create(ctx, dataset, &honeycombio.Column{
		KeyName:     "column_name",
		Description: "This column is created by dd resource test",
	})
	if err != nil {
		t.Error(err)
	}

	//Josslyn : TODO this needs some additional work to refresh after the plan gets updated, currently doesn't pass
	resource.Test(t, resource.TestCase{
		PreCheck:          testAccPreCheck(t),
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				// initial setup with no dataset definitons
				Config: testAccDatasetDefinitonConfig(dataset),
			},
			{
				// update the name field to a non-default value
				Config: fmt.Sprintf(`
				resource "honeycombio_dataset_definition" "test" {
				  dataset     = "%s"
				  field {
					name = "name"
					value = "%s"
					}
				}`, dataset, col.KeyName),

				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatasetDefinitionExists(t, dataset, "honeycombio_dataset_definition.test", "name"),
					resource.TestCheckResourceAttr("honeycombio_dataset_definition.test", "field.0.name", "name"),
					resource.TestCheckResourceAttr("honeycombio_dataset_definition.test", "field.0.value", "column_name"),
				),
			},
			// {
			// 	// reset the name field to the default value by ommitting it
			// 	// set duration_ms to a non-default field
			// 	Config: `
			// 	resource "honeycombio_dataset_definition" "test" {
			// 	dataset    = "testacc"

			// 	field {
			// 		name = "duration_ms"
			// 		value = "alternate_durationMs"
			// 	}
			// 		}`,

			// 	Check: resource.ComposeTestCheckFunc(
			// 		resource.TestCheckResourceAttr("honeycombio_dataset_definition.test", "field.0.name", "duration_ms"),
			// 		resource.TestCheckResourceAttr("honeycombio_dataset_definition.test", "field.0.value", "alternate_durationMs"),
			// 	),
			// },
			// { // update duration ms field to existing derived column
			// 	Config: `
			// 		resource "honeycombio_dataset_definition" "test" {
			// 		dataset    = "testacc"

			// 		field {
			// 			name = "duration_ms"
			// 			value = "gt50_duration_ms"
			// 		}
			// 			}`,

			// 	Check: resource.ComposeTestCheckFunc(
			// 		resource.TestCheckResourceAttr("honeycombio_dataset_definition.test", "field.0.name", "duration_ms"),
			// 		resource.TestCheckResourceAttr("honeycombio_dataset_definition.test", "field.0.value", "gt50_duration_ms"),
			// 	),
			// },
			// { // update all dataset definitions to non-default fields
			// 	// Check that setting every dataset defintion works as expected
			// 	Config: `
			// 		resource "honeycombio_dataset_definition" "test" {
			// 		dataset    = "testacc"

			// 		field {
			// 		  name = "name"
			// 		  value = "pr_branch"
			// 		}

			// 		field {
			// 			name = "duration_ms"
			// 			value = "gt50_duration_ms"
			// 		}

			// 		field {
			// 			  name = "parent_id"
			// 			  value = "branch"
			// 		}

			// 		field {
			// 			  name = "service_name"
			// 			  value = "build_num"
			// 		}

			// 		field {
			// 			  name = "trace_id"
			// 			  value = "ci_provider"
			// 		}

			// 		field {
			// 			  name = "span_id"
			// 			  value = "github.sha"
			// 		}

			// 		field {
			// 			name = "error"
			// 			value = "github.workflow"
			// 	  	}

			// 		field {
			// 			name = "route"
			// 			value = "command_name"
			// 	  	}

			// 		field {
			// 			name = "span_kind"
			// 			value = "github.actor"
			// 	  	}

			// 		field {
			// 			name = "annotation_type"
			// 			value = "github.base_ref"
			// 	  	}

			// 		  field {
			// 			name = "link_trace_id"
			// 			value = "github.event_name"
			// 	  	}

			// 		field {
			// 			name = "link_span_id"
			// 			value = "github.head_ref"
			// 	  	}

			// 		  field {
			// 			name = "status"
			// 			value = "job.status"
			// 	  	}

			// 		field {
			// 			name = "user"
			// 			value = "meta.source"
			// 	  	}

			// 	}`,
			// test will call destroy after this step which
			// invokes Delete which resets everything back to default values
			// },
		},
	})
}

func testAccDatasetDefinitonConfig(dataset string) string {
	return fmt.Sprintf(`
resource "honeycombio_dataset_definition" "test" {
  dataset = "%s"
}`, dataset)
}

func testAccCheckDatasetDefinitionExists(t *testing.T, dataset string, resourceName string, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resourceState, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		client := testAccClient(t)
		createdDef, err := client.DatasetDefinitions.Get(context.Background(), dataset)
		if err != nil {
			return fmt.Errorf("could not find created dataset definition: %w", err)
		}

		assert.Equal(t, name, createdDef.Name)
		assert.Equal(t, resourceState.Primary.Attributes["DurationMs"], createdDef.DurationMs.Name)

		return nil
	}
}
