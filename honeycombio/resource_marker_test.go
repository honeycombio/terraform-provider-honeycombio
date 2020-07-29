package honeycombio

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	honeycombio "github.com/kvrhdn/go-honeycombio"
)

func TestAccHoneycombioMarker_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccMarkerConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMarkerExists("honeycombio_marker.test"),
					resource.TestCheckResourceAttr("honeycombio_marker.test", "message", "Hello world!"),
					resource.TestCheckResourceAttr("honeycombio_marker.test", "type", "deploys"),
					resource.TestCheckResourceAttr("honeycombio_marker.test", "url", "https://www.honeycomb.io/"),
				),
			},
		},
	})
}

func testAccMarkerConfig() string {
	return `
resource "honeycombio_marker" "test" {
  message = "Hello world!"
  type    = "deploys"
  url     = "https://www.honeycomb.io/"
}`
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
		existingMarker, err := client.Markers.Get(resourceState.Primary.ID)
		if err != nil {
			return fmt.Errorf("could not retrieve marker: %w", err)
		}

		if existingMarker.Message != "Hello world!" {
			return fmt.Errorf("bad active state, expected message = \"Hello world!\", got: %s", existingMarker.Message)
		}
		if existingMarker.Type != "deploys" {
			return fmt.Errorf("bad active state, expected type = \"deploys\", got: %s", existingMarker.Type)
		}
		if existingMarker.URL != "https://www.honeycomb.io/" {
			return fmt.Errorf("bad active state, expected url = \"https://www.honeycomb.io/\", got: %s", existingMarker.URL)
		}

		return nil
	}
}
