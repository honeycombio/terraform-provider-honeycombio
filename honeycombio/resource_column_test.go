package honeycombio

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccHoneycombioColumn_basic(t *testing.T) {
	dataset := testAccDataset()
	keyName := acctest.RandomWithPrefix("duration_ms_test")

	resource.Test(t, resource.TestCase{
		PreCheck:          testAccPreCheck(t),
		ProviderFactories: testAccProviderFactories,
		IDRefreshName:     "honeycombio_column.test",
		Steps: []resource.TestStep{
			{
				Config: testAccColumnConfig(keyName, dataset),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("honeycombio_column.test", "key_name", keyName),
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

func testAccColumnConfig(keyName, dataset string) string {
	return fmt.Sprintf(`
resource "honeycombio_column" "test" {
  key_name    = "%s"
  type        = "float"
  hidden      = false
  description = "Duration of the trace"

  dataset = "%s"
}`, keyName, dataset)
}

func TestAccHoneycombioColumn_validationErrors(t *testing.T) {
	dataset := testAccDataset()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          testAccPreCheck(t),
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccColumnConfigWithType(dataset, `String`),
				ExpectError: regexp.MustCompile("expected type to be one of \\[string float integer boolean\\], got String"),
			},
		},
	})
}

func testAccColumnConfigWithType(dataset, typeStr string) string {
	return fmt.Sprintf(`
resource "honeycombio_column" "test" {
  key_name = "duration_ms_test"
  type     = "%s"

  dataset = "%s"
}`, typeStr, dataset)
}
