package honeycombio

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	honeycombio "github.com/kvrhdn/go-honeycombio"
)

func TestAccHoneycombioMarker_basic(t *testing.T) {
	dataset := testAccDataset()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccMarkerConfig(dataset),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMarkerExists("honeycombio_marker.test"),
					resource.TestCheckResourceAttr("honeycombio_marker.test", "message", "Hello world!"),
					resource.TestCheckResourceAttr("honeycombio_marker.test", "type", "deploys"),
					resource.TestCheckResourceAttr("honeycombio_marker.test", "url", "https://www.honeycomb.io/"),
					resource.TestCheckResourceAttr("honeycombio_marker.test", "dataset", dataset),
				),
			},
		},
	})
}

func testAccMarkerConfig(dataset string) string {
	return fmt.Sprintf(`
resource "honeycombio_marker" "test" {
  message = "Hello world!"
  type    = "deploys"
  url     = "https://www.honeycomb.io/"
  dataset = "%s"
}`, dataset)
}

// testAccCheckMarkerExists queries the API to verify the Marker exists and
// matches with the Terraform state.
func testAccCheckMarkerExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resourceState, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		client := testAccProvider.Meta().(*honeycombio.Client)
		_, err := client.Markers.Get(context.Background(), resourceState.Primary.Attributes["dataset"], resourceState.Primary.ID)
		if err != nil {
			return fmt.Errorf("could not retrieve marker: %w", err)
		}

		return nil
	}
}
