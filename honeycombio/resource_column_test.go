package honeycombio

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccHoneycombioColumn_basic(t *testing.T) {
	dataset := testAccDataset()
	keyName := acctest.RandomWithPrefix("duration_ms_test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
		IDRefreshName:            "honeycombio_column.test",
		Steps: []resource.TestStep{
			{
				Config: testAccColumnConfig(keyName, dataset),
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
		CheckDestroy: resource.ComposeTestCheckFunc(
			func(s *terraform.State) error {
				// ensure column is removed on destroy
				client := testAccClient(t)
				_, err := client.Columns.GetByKeyName(context.Background(), dataset, keyName)
				if err == nil {
					return fmt.Errorf("column %q was not deleted on destroy", keyName)
				}
				return nil
			},
		),
	})
}

func testAccColumnConfig(keyName, dataset string) string {
	return fmt.Sprintf(`
resource "honeycombio_column" "test" {
  name        = "%s"
  type        = "float"
  hidden      = false
  description = "Duration of the trace"

  dataset = "%s"
}`, keyName, dataset)
}
