package honeycombio

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
)

func TestAccHoneycombioDatasetDefinition_basic(t *testing.T) {
	var definitionBefore, definitionAfter honeycombio.DatasetDefinition

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
					//
					testAccCheckDatasetDefinitionExistsInHoneycomb(t, "honeycombio_dataset_definition.test", &definitionBefore),
					testAccCheckDatasetDefinitionAttributes(&definitionBefore),

					resource.TestCheckResourceAttr("honeycombio_dataset_definition.test", "field.0.name", "duration_ms"),
					resource.TestCheckResourceAttr("honeycombio_dataset_definition.test", "field.0.value", "duration_ms"),
				),
			},
			{
				// change the name field
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
					//
					testAccCheckDatasetDefinitionExistsInHoneycomb(t, "honeycombio_dataset_definition.test", &definitionAfter),
					testAccCheckDatasetDefinitionAttributes(&definitionAfter),

					resource.TestCheckResourceAttr("honeycombio_dataset_definition.test", "field.0.name", "duration_ms"),
					resource.TestCheckResourceAttr("honeycombio_dataset_definition.test", "field.0.value", "duration_ms"),
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
					//
					testAccCheckDatasetDefinitionExistsInHoneycomb(t, "honeycombio_dataset_definition.test", &definitionBefore),
					testAccCheckDatasetDefinitionAttributes(&definitionBefore),

					resource.TestCheckResourceAttr("honeycombio_dataset_definition.test", "field.0.name", "duration_ms"),
					resource.TestCheckResourceAttr("honeycombio_dataset_definition.test", "field.0.value", "duration_ms"),
				),
			},
		},
	})
}

func testAccCheckDatasetDefinitionExistsInHoneycomb(t *testing.T, name string, dd *honeycombio.DatasetDefinition) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resourceState, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		client := testAccClient(t)
		createdDatasetDefinition, err := client.DatasetDefinitions.Get(context.Background(), resourceState.Primary.Attributes["dataset"])
		if err != nil {
			return fmt.Errorf("could not find created definition: %w", err)
		}

		*dd = *createdDatasetDefinition
		return nil
	}
}

func testAccCheckDatasetDefinitionAttributes(dd *honeycombio.DatasetDefinition) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if dd.TraceID.Name != "trace.trace_id" {
			return fmt.Errorf("bad name: %s", dd.TraceID.Name)
		}

		if dd.TraceID.ColumnType != "column" {
			return fmt.Errorf("bad column_type: %s", dd.TraceID.Name)
		}

		return nil
	}
}

func testAccDatasetDefinition(dataset string, dName0 string, dValue0 string) string {
	return fmt.Sprintf(`
resource "honeycombio_dataset_definition" "test" {
  dataset    = "%s"
  
  field {
	name = "%s"
	value = "%s"
  }
}`, dataset, dName0, dValue0)
}

func testAccDatasetDefinitionThree(dataset string, dName0 string, dValue0 string, dName1 string, dValue1 string, dName2 string, dValue2 string) string {
	return fmt.Sprintf(`
resource "honeycombio_dataset_definition" "test" {
  dataset    = "%s"
  
  field {
	name = "%s"
	value = "%s"
  }

  field {
	name = "%s"
	value = "%s"
  }

  field {
	name = "%s"
	value = "%s"
  }
}`, dataset, dName0, dValue0, dName1, dValue1, dName2, dValue2)
}
