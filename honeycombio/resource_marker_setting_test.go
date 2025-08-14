package honeycombio

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"

	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
)

func TestAccHoneycombioMarkerSetting_basic(t *testing.T) {
	dataset := testAccDataset()

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "honeycombio_marker_setting" "test" {
  color   = "#7b1fa2"
  type    = "test123"
  dataset = "%s"
}`, dataset),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMarkerSettingExists(t, "honeycombio_marker_setting.test", dataset),
					resource.TestCheckResourceAttr("honeycombio_marker_setting.test", "color", "#7b1fa2"),
					resource.TestCheckResourceAttr("honeycombio_marker_setting.test", "type", "test123"),
					resource.TestCheckResourceAttr("honeycombio_marker_setting.test", "dataset", dataset),
				),
			},
		},
	})
}

func TestAccHoneycombioMarkerSetting_AllToUnset(t *testing.T) {
	ctx := t.Context()
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
resource "honeycombio_marker_setting" "test" {
  color   = "#000000"	
  type    = "testy"
  dataset = "__all__"
}`,
				Check: testAccCheckMarkerSettingExists(t, "honeycombio_marker_setting.test", honeycombio.EnvironmentWideSlug),
			},
			{
				Config: `
resource "honeycombio_marker_setting" "test" {
  color = "#000000"	
  type  = "testy"
}`,
				Check:              testAccCheckMarkerSettingExists(t, "honeycombio_marker_setting.test", honeycombio.EnvironmentWideSlug),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccCheckMarkerSettingExists(t *testing.T, name, dataset string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resourceState, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		c := testAccClient(t)
		m, err := c.MarkerSettings.Get(t.Context(), dataset, resourceState.Primary.ID)
		if err != nil {
			return fmt.Errorf("could not retrieve marker settings: %w", err)
		}

		assert.Equal(t, resourceState.Primary.ID, m.ID)

		return nil
	}
}
