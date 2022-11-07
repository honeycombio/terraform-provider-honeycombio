package honeycombio

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccHoneycombioDatasetDefinition_basic(t *testing.T) {
	// set multiple definitions in a single HCL block
	dataset := testAccDataset()

	resource.Test(t, resource.TestCase{
		PreCheck:          testAccPreCheck(t),
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				// initial setup, don't set anything, all default values
				// MOLLY: I don't understand why this step is necessary
				// but tests won't pass without it
				Config: testAccDatasetDefinitonConfig(dataset),
			},
			// {
			// 	// update the name field to a non-default value
			// 	Config: `
			// 	resource "honeycombio_dataset_definition" "test" {
			// 	dataset    = "testacc"

			// 	field {
			// 		name = "name"
			// 		value = "job.status"
			// 	}
			// 		}`,
			// 	Check: resource.ComposeTestCheckFunc(
			// 		resource.TestCheckResourceAttr("honeycombio_dataset_definition.test", "field.0.name", "name"),
			// 		resource.TestCheckResourceAttr("honeycombio_dataset_definition.test", "field.0.value", "job.status"),
			// 	),
			// },
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
