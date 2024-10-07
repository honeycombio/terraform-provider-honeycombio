package honeycombio

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/test"
)

func TestAccHoneycombioColumn_basic(t *testing.T) {
	dataset := testAccDataset()
	keyName := test.RandomStringWithPrefix("test.", 10)

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
		IDRefreshName:            "honeycombio_column.test",
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "honeycombio_column" "test" {
  name        = "%s"
  type        = "float"
  hidden      = false
  description = "Duration of the trace"

  dataset = "%s"
}`, keyName, dataset),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("honeycombio_column.test", "name", keyName),
					resource.TestCheckResourceAttr("honeycombio_column.test", "type", "float"),
					resource.TestCheckResourceAttr("honeycombio_column.test", "hidden", "false"),
					resource.TestCheckResourceAttr("honeycombio_column.test", "description", "Duration of the trace"),
				),
			},
			{
				ResourceName:      "honeycombio_column.test",
				ImportStateId:     fmt.Sprintf("%s/%s", dataset, keyName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
