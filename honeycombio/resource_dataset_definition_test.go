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
	var definitionBefore honeycombio.DatasetDefinition

	dataset := testAccDataset()

	// set multiple definitions in a single HCL block

	resource.Test(t, resource.TestCase{
		PreCheck:          testAccPreCheck(t),
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasetDefinition(dataset, "trace_id", "trace.hc_terraform"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("honeycombio_dataset_definition.test", "dataset", dataset),
					resource.TestCheckResourceAttr("honeycombio_dataset_definition.test", "field.0.name", "trace_id"),
					resource.TestCheckResourceAttr("honeycombio_dataset_definition.test", "field.0.value", "trace.hc_terraform"),
					//Config: testAccDatasetDefinitionThree(dataset, "trace_id", "trace.hc_terraform", "error", "error.hc_terraform", "status", "status.hc_terraform"),
					//resource.TestCheckResourceAttr("honeycombio_dataset_definition.test", "field.1.name", "error"),
					//resource.TestCheckResourceAttr("honeycombio_dataset_definition.test", "field.1.value", "error.hc_terraform"),
					//resource.TestCheckResourceAttr("honeycombio_dataset_definition.test", "field.2.name", "status"),
					//resource.TestCheckResourceAttr("honeycombio_dataset_definition.test", "field.2.value", "status.hc_terraform"),
					testAccCheckDatasetDefinitionExists(t, "honeycombio_dataset_definition.test", &definitionBefore),
					testAccCheckDatasetDefinitionAttributes(&definitionBefore),
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
