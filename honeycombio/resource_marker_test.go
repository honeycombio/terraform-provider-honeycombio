package honeycombio

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
)

func TestAccHoneycombioMarker_basic(t *testing.T) {
	dataset := testAccDataset()

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactory,
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

func TestAccHoneycombioMarker_timeRange(t *testing.T) {
	dataset := testAccDataset()

	startTime := 1700000000
	endTime := startTime + 3600

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactory,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "honeycombio_marker" "test" {
  message    = "time-range marker"
  type       = "deploy"
  start_time = %d
  end_time   = %d
  dataset    = "%s"
}`, startTime, endTime, dataset),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMarkerExists(t, "honeycombio_marker.test", dataset),
					resource.TestCheckResourceAttr("honeycombio_marker.test", "start_time", fmt.Sprintf("%d", startTime)),
					resource.TestCheckResourceAttr("honeycombio_marker.test", "end_time", fmt.Sprintf("%d", endTime)),
				),
			},
		},
	})
}

func TestAccHoneycombioMarker_endBeforeStart(t *testing.T) {
	dataset := testAccDataset()

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactory,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "honeycombio_marker" "test" {
  message    = "invalid range"
  type       = "deploy"
  start_time = 1700003600
  end_time   = 1700000000
  dataset    = "%s"
}`, dataset),
				ExpectError: regexp.MustCompile(`end_time .* must be greater than or equal to start_time`),
			},
		},
	})
}

func TestAccHoneycombioMarker_endWithoutStart(t *testing.T) {
	dataset := testAccDataset()

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactory,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "honeycombio_marker" "test" {
  message  = "end without start"
  type     = "deploy"
  end_time = 1700000000
  dataset  = "%s"
}`, dataset),
				ExpectError: regexp.MustCompile(`end_time cannot be set without start_time`),
			},
		},
	})
}

func TestAccHoneycombioMarker_startTimeComputed(t *testing.T) {
	dataset := testAccDataset()

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactory,
		Steps: []resource.TestStep{
			{
				// start_time is omitted: the API defaults it to "now" and the
				// Computed attribute must absorb that value without drift.
				Config: fmt.Sprintf(`
resource "honeycombio_marker" "test" {
  message = "no explicit start_time"
  type    = "deploy"
  dataset = "%s"
}`, dataset),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMarkerExists(t, "honeycombio_marker.test", dataset),
					resource.TestCheckResourceAttrSet("honeycombio_marker.test", "start_time"),
				),
			},
			{
				// Re-planning with the same config must produce no diff.
				Config: fmt.Sprintf(`
resource "honeycombio_marker" "test" {
  message = "no explicit start_time"
  type    = "deploy"
  dataset = "%s"
}`, dataset),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
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
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactory,
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
