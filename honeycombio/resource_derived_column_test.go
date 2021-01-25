package honeycombio

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccHoneycombioDerivedColumn_basic(t *testing.T) {
	dataset := testAccDataset()

	resource.Test(t, resource.TestCase{
		PreCheck:          testAccPreCheck(t),
		ProviderFactories: testAccProviderFactories,
		IDRefreshName:     "honeycombio_derived_column.test",
		Steps: []resource.TestStep{
			{
				Config: testAccDerivedColumnConfig(dataset),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("honeycombio_derived_column.test", "alias", "duration_ms_log10"),
					resource.TestCheckResourceAttr("honeycombio_derived_column.test", "expression", "LOG10($duration_ms)"),
					resource.TestCheckResourceAttr("honeycombio_derived_column.test", "description", "LOG10 of duration_ms"),
				),
			},
			{
				ResourceName:      "honeycombio_derived_column.test",
				ImportStateId:     fmt.Sprintf("%s/%s", dataset, "duration_ms_log10"),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccDerivedColumnConfig(dataset string) string {
	return fmt.Sprintf(`
resource "honeycombio_derived_column" "test" {
  alias       = "duration_ms_log10"
  expression  = "LOG10($duration_ms)"
  description = "LOG10 of duration_ms"

  dataset = "%s"
}`, dataset)
}
