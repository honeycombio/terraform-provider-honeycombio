package honeycombio

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
)

func TestAccHoneycombioMarkerSetting_basic(t *testing.T) {
	dataset := testAccDataset()

	resource.Test(t, resource.TestCase{
		PreCheck:          testAccPreCheck(t),
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMarkerSettingConfig(dataset),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMarkerSettingExists(t, "honeycombio_marker_setting.test"),
					resource.TestCheckResourceAttr("honeycombio_marker_setting.test", "color", "#7b1fa2"),
					resource.TestCheckResourceAttr("honeycombio_marker_setting.test", "type", "test"),
					resource.TestCheckResourceAttr("honeycombio_marker_setting.test", "dataset", dataset),
				),
			},
		},
	})
}

func testAccMarkerSettingConfig(dataset string) string {
	return fmt.Sprintf(`
resource "honeycombio_marker_setting" "test" {
  color = "#7b1fa2"
  type    = "test"
  dataset = "%s"
}`, dataset)
}

// testAccCheckMarkerSettingExists queries the API to verify the Marker Setting exists and
// matches with the Terraform state.
func testAccCheckMarkerSettingExists(t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resourceState, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		c := testAccClient(t)
		m, err := c.MarkerSettings.Get(context.Background(), resourceState.Primary.Attributes["dataset"], resourceState.Primary.ID)
		if err != nil {
			return fmt.Errorf("could not retrieve marker settings: %w", err)
		}

		assert.Equal(t, "test", m.Type)

		return nil
	}
}
