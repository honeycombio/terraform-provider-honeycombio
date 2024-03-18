package honeycombio

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccHoneycombioQueryAnnotation_update(t *testing.T) {
	dataset := testAccDataset()
	firstName := "first annotation name"
	secondName := "second annotation name"

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceQueryAnnotationConfig(dataset, firstName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueryAnnotationExists(t, dataset, "honeycombio_query_annotation.test"),
					resource.TestCheckResourceAttr("honeycombio_query_annotation.test", "name", firstName),
					resource.TestCheckResourceAttr("honeycombio_query_annotation.test", "description", "Test query annotation description"),
					resource.TestCheckResourceAttrPair("honeycombio_query_annotation.test", "query_id", "honeycombio_query.test", "id"),
				),
			},
			{
				Config: testAccResourceQueryAnnotationConfig(dataset, secondName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueryAnnotationExists(t, dataset, "honeycombio_query_annotation.test"),
					resource.TestCheckResourceAttr("honeycombio_query_annotation.test", "name", secondName),
					resource.TestCheckResourceAttr("honeycombio_query_annotation.test", "description", "Test query annotation description"),
					resource.TestCheckResourceAttrPair("honeycombio_query_annotation.test", "query_id", "honeycombio_query.test", "id"),
				),
			},
		},
	})
}

func testAccResourceQueryAnnotationConfig(dataset string, name string) string {
	return fmt.Sprintf(`
data "honeycombio_query_specification" "test" {
  calculation {
    op     = "AVG"
    column = "duration_ms"
  }

  filter {
    column = "duration_ms"
    op     = ">"
    value  = 10
  }
}

resource "honeycombio_query" "test" {
  dataset = "%s"
  query_json = data.honeycombio_query_specification.test.json
}

resource "honeycombio_query_annotation" "test" {
	dataset = "%s"
	query_id = honeycombio_query.test.id
	name = "%s"
	description = "Test query annotation description"
}
`, dataset, dataset, name)
}

func testAccCheckQueryAnnotationExists(t *testing.T, dataset string, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resourceState, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		client := testAccClient(t)
		_, err := client.QueryAnnotations.Get(context.Background(), dataset, resourceState.Primary.ID)
		if err != nil {
			return fmt.Errorf("could not find created query: %w", err)
		}

		return nil
	}
}
