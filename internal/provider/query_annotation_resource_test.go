package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
)

func TestAcc_HoneycombioQueryAnnotation(t *testing.T) {
	dataset := testAccDataset()
	firstName := "first annotation name"
	secondName := "second annotation name"

	config := `
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
}`

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(config, dataset, dataset, firstName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueryAnnotationExists(t, dataset, "honeycombio_query_annotation.test"),
					resource.TestCheckResourceAttr("honeycombio_query_annotation.test", "name", firstName),
					resource.TestCheckResourceAttr("honeycombio_query_annotation.test", "description", "Test query annotation description"),
					resource.TestCheckResourceAttrPair("honeycombio_query_annotation.test", "query_id", "honeycombio_query.test", "id"),
				),
			},
			{
				Config: fmt.Sprintf(config, dataset, dataset, secondName),
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

func TestAcc_HoneycombioQueryAnnotation_AllToUnset(t *testing.T) {
	ctx := context.Background()
	c := testAccClient(t)

	if c.IsClassic(ctx) {
		t.Skip("env-wide query annotations are not supported in classic")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
		Steps: []resource.TestStep{
			{
				Config: `
resource "honeycombio_query" "test" {
  query_json = "{}"
}

resource "honeycombio_query_annotation" "test" {
  dataset     = "__all__"
  query_id    = honeycombio_query.test.id
	name        = "test annotation"
	description = "Test query annotation description"
}`,
				Check: testAccCheckQueryAnnotationExists(t, client.EnvironmentWideSlug, "honeycombio_query_annotation.test"),
			},
			{
				Config: `
resource "honeycombio_query" "test" {
  query_json = "{}"
}

resource "honeycombio_query_annotation" "test" {
  query_id    = honeycombio_query.test.id
	name        = "test annotation"
	description = "Test query annotation description"
}`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// TestAcc_QueryAnnotationResource_UpgradeFromVersion0381 tests the migration case from the
// last SDK-based version of the Query Annotation resource to the current Framework-based version.
//
// See: https://developer.hashicorp.com/terraform/plugin/framework/migrating/testing#testing-migration
func TestAcc_QueryAnnotationResource_UpgradeFromVersion0381(t *testing.T) {
	dataset := testAccDataset()
	name := "test annotation name"
	config := fmt.Sprintf(`
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
}`, dataset, dataset, name)

	resource.Test(t, resource.TestCase{
		PreCheck: testAccPreCheck(t),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"honeycombio": {
						VersionConstraint: "0.38.1",
						Source:            "honeycombio/honeycombio",
					},
				},
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckQueryAnnotationExists(t, dataset, "honeycombio_query_annotation.test"),
				),
			},
			{
				ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
				Config:                   config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("honeycombio_query_annotation.test", "id"),
					resource.TestCheckResourceAttr("honeycombio_query_annotation.test", "name", name),
					resource.TestCheckResourceAttr("honeycombio_query_annotation.test", "description", "Test query annotation description"),
					resource.TestCheckResourceAttrPair("honeycombio_query_annotation.test", "query_id", "honeycombio_query.test", "id"),
				),
			},
		},
	})
}

func testAccCheckQueryAnnotationExists(t *testing.T, dataset, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resourceState, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		client := testAccClient(t)
		_, err := client.QueryAnnotations.Get(context.Background(), dataset, resourceState.Primary.ID)
		if err != nil {
			return fmt.Errorf("could not find created query annotation: %w", err)
		}

		return nil
	}
}
