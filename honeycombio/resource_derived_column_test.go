package honeycombio

import (
	"fmt"
	"regexp"
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
				Config: testAccDerivedColumnConfig(dataset, "duration_ms_log10"),
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

	// validate 'pretty' alias'
	resource.Test(t, resource.TestCase{
		PreCheck:          testAccPreCheck(t),
		ProviderFactories: testAccProviderFactories,
		IDRefreshName:     "honeycombio_derived_column.test",
		Steps: []resource.TestStep{
			{
				Config: testAccDerivedColumnConfig(dataset, "LOG(10) duration_ms"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("honeycombio_derived_column.test", "alias", "LOG(10) duration_ms"),
					resource.TestCheckResourceAttr("honeycombio_derived_column.test", "expression", "LOG10($duration_ms)"),
					resource.TestCheckResourceAttr("honeycombio_derived_column.test", "description", "LOG10 of duration_ms"),
				),
			},
		},
	})
}

func TestAccHoneycombioDerivedColumn_error(t *testing.T) {
	dataset := testAccDataset()

	resource.Test(t, resource.TestCase{
		PreCheck:          testAccPreCheck(t),
		ProviderFactories: testAccProviderFactories,
		IDRefreshName:     "honeycombio_derived_column.test",
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "honeycombio_derived_column" "invalid_column_in_expression" {
  alias       = "invalid_column_in_expression"
  expression  = "LOG10($invalid_column)"

  dataset = "%s"
}
`, dataset),
				ExpectError: regexp.MustCompile(`Error: unknown column name: invalid_column`),
			},
		},
	})
}

func testAccDerivedColumnConfig(dataset, alias string) string {
	return fmt.Sprintf(`
resource "honeycombio_derived_column" "test" {
  alias       = "%s"
  expression  = "LOG10($duration_ms)"
  description = "LOG10 of duration_ms"

  dataset = "%s"
}`, alias, dataset)
}
