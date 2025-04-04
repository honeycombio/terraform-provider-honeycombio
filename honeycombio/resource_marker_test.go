package honeycombio

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
)

func TestAccHoneycombioMarker_basic(t *testing.T) {
	dataset := testAccDataset()

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "honeycombio_marker" "test" {
  message = "Hello world!"
  type    = "deploy"
  url     = "https://www.honeycomb.io/"
  dataset = "%s"
}`, dataset),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMarkerExists(t, "honeycombio_marker.test", dataset),
					resource.TestCheckResourceAttr("honeycombio_marker.test", "message", "Hello world!"),
					resource.TestCheckResourceAttr("honeycombio_marker.test", "type", "deploy"),
					resource.TestCheckResourceAttr("honeycombio_marker.test", "url", "https://www.honeycomb.io/"),
					resource.TestCheckResourceAttr("honeycombio_marker.test", "dataset", dataset),
				),
			},
		},
	})
}

func TestAccHoneycombioMarker_AllToUnset(t *testing.T) {
	ctx := context.Background()
	c := testAccClient(t)

	if c.IsClassic(ctx) {
		t.Skip("env-wide markers are not supported in classic")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
		Steps: []resource.TestStep{
			{
				Config: `
resource "honeycombio_marker" "test" {		
  message = "hey"
	type    = "test"
  dataset = "__all__"
}`,
				Check: testAccCheckMarkerExists(t, "honeycombio_marker.test", honeycombio.EnvironmentWideSlug),
			},
			{
				Config: `
resource "honeycombio_marker" "test" {		
  message = "hey"
	type    = "test"
}`,
				Check:              testAccCheckMarkerExists(t, "honeycombio_marker.test", honeycombio.EnvironmentWideSlug),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccCheckMarkerExists(t *testing.T, name, dataset string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resourceState, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		c := testAccClient(t)
		_, err := c.Markers.Get(context.Background(), dataset, resourceState.Primary.ID)
		if err != nil {
			return fmt.Errorf("could not retrieve marker: %w", err)
		}

		return nil
	}
}
