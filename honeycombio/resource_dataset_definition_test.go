package honeycombio

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccHoneycombioDatasetDefinition_basic(t *testing.T) {
	// dataset := testAccDataset()

	// set multiple definitions in a single HCL block
	resource.Test(t, resource.TestCase{
		PreCheck:          testAccPreCheck(t),
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				// initial setup
				Config: `
				resource "honeycombio_dataset_definition" "test" {
				dataset    = "testacc"
				
				field {
					name = "duration_ms"
					value = "duration_ms"
				}

				field {
					name = "name"
					value = "name"
				}

				field {
					name = "parent_id"
					value = "trace.parent_id"
				}

				field {
					name = "service_name"
					value = "service_name"
				}

				field {
					name = "trace_id"
					value = "trace.trace_id"
				}

				field {
					name = "span_id"
					value = "trace.span_id"
				}
					}`,

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("honeycombio_dataset_definition.test", "field.0.name", "duration_ms"),
					resource.TestCheckResourceAttr("honeycombio_dataset_definition.test", "field.0.value", "duration_ms"),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				// update the name field
				Config: `
				resource "honeycombio_dataset_definition" "test" {
				dataset    = "testacc"

				field {
					name = "duration_ms"
					value = "duration_ms"
				}

				field {
					name = "name"
					value = "job.status"
				}

				field {
					name = "parent_id"
					value = "trace.parent_id"
				}

				field {
					name = "service_name"
					value = "service_name"
				}

				field {
					name = "trace_id"
					value = "trace.trace_id"
				}

				field {
					name = "span_id"
					value = "trace.span_id"
				}
					}`,

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("honeycombio_dataset_definition.test", "field.1.name", "name"),
					resource.TestCheckResourceAttr("honeycombio_dataset_definition.test", "field.1.value", "job.status"),
				),
			},
			{ // reset the name field to its original values
				Config: `
				resource "honeycombio_dataset_definition" "test" {
				dataset    = "testacc"
				
				field {
					name = "duration_ms"
					value = "duration_ms"
				}

				field {
					name = "name"
					value = "name"
				}

				field {
					name = "parent_id"
					value = "trace.parent_id"
				}

				field {
					name = "service_name"
					value = "service_name"
				}

				field {
					name = "trace_id"
					value = "trace.trace_id"
				}

				field {
					name = "span_id"
					value = "trace.span_id"
				}
					}`,

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("honeycombio_dataset_definition.test", "field.1.name", "name"),
					resource.TestCheckResourceAttr("honeycombio_dataset_definition.test", "field.1.value", "name"),
				),
			},
			{ // update duration ms field to existing derived column
				Config: `
					resource "honeycombio_dataset_definition" "test" {
					dataset    = "testacc"
					

					field {
						name = "duration_ms"
						value = "gt50_duration_ms"
					}

					field {
						name = "parent_id"
						value = "trace.parent_id"
					}

					field {
						name = "service_name"
						value = "service_name"
					}

					field {
						name = "trace_id"
						value = "trace.trace_id"
					}

					field {
						name = "span_id"
						value = "trace.span_id"
					}
						}`,

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("honeycombio_dataset_definition.test", "field.0.name", "duration_ms"),
					resource.TestCheckResourceAttr("honeycombio_dataset_definition.test", "field.0.value", "gt50_duration_ms"),
				),
				ExpectNonEmptyPlan: true,
			},
			{ // delete the name field
				Config: `
				resource "honeycombio_dataset_definition" "test" {
				dataset    = "testacc"
				
				field {
					name = "duration_ms"
					value = "gt50_duration_ms"
				}

				field {
					name = "parent_id"
					value = "trace.parent_id"
				}

				field {
					name = "service_name"
					value = "service_name"
				}

				field {
					name = "trace_id"
					value = "trace.trace_id"
				}

				field {
					name = "span_id"
					value = "trace.span_id"
				}
					}`,

				// TODO: find name field, what would HCL return
				Check: resource.ComposeTestCheckFunc(
				//resource.TestCheckTypeSetElemAttr("honeycombio_dataset_definition.test", "field.0.name", ""),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				// reset to default
				Config: `
				resource "honeycombio_dataset_definition" "test" {
				dataset    = "testacc"
				
				field {
					name = "duration_ms"
					value = "duration_ms"
				}

				field {
					name = "name"
					value = "name"
				}

				field {
					name = "parent_id"
					value = "trace.parent_id"
				}

				field {
					name = "service_name"
					value = "service_name"
				}

				field {
					name = "trace_id"
					value = "trace.trace_id"
				}

				field {
					name = "span_id"
					value = "trace.span_id"
				}
					}`,

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("honeycombio_dataset_definition.test", "field.0.name", "duration_ms"),
					resource.TestCheckResourceAttr("honeycombio_dataset_definition.test", "field.0.value", "duration_ms"),
				),
			},
		},
	})
}
