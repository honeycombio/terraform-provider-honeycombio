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
	//var columnBefore, columnAfter honeycombio.DefinitionColumn
	var definitionBefore, definitionAfter honeycombio.DatasetDefinition

	dataset := testAccDataset()

	resource.Test(t, resource.TestCase{
		PreCheck:          testAccPreCheck(t),
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasetDefinitionWithTraceID(dataset, "trace.trace_id"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("honeycombio_dataset_definition.test", "dataset", dataset),
					resource.TestCheckResourceAttr("honeycombio_dataset_definition.test", "trace_id", dataset),
					testAccCheckDatasetDefinitionExists(t, "honeycombio_dataset_definition.test", &definitionBefore),
					testAccCheckDatasetDefinitionAttributes(&definitionBefore),
				),
			},
			{
				Config: testAccDatasetDefinitionWithTraceID(dataset, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("honeycombio_dataset_definition.test", "dataset", dataset),
					resource.TestCheckResourceAttr("honeycombio_dataset_definition.test", "trace_id", dataset),
					testAccCheckDatasetDefinitionExists(t, "honeycombio_dataset_definition.test", &definitionAfter),
					testAccCheckDatasetDefinitionAttributes(&definitionAfter),
				),
			},
		},
	})
}

func testAccCheckDatasetDefinitionExists(t *testing.T, name string, dd *honeycombio.DatasetDefinition) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resourceState, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		client := testAccClient(t)
		createdDatasetDefinition, err := client.DatasetDefinitions.List(context.Background(), resourceState.Primary.Attributes["dataset"])
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

func testAccDatasetDefinitionWithTraceID(dataset string, definitionValue string) string {
	return fmt.Sprintf(`
resource "honeycombio_dataset_definition" "test" {
  dataset    = "%s"
  
  trace_id {
	name = "%s"
  }
}`, dataset, definitionValue)
}
