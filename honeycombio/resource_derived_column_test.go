package honeycombio

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/test"
)

func TestAccHoneycombioDerivedColumn_basic(t *testing.T) {
	dataset := testAccDataset()
	alias := test.RandomStringWithPrefix("test.", 10)

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "honeycombio_derived_column" "test" {
  alias       = "%s"
  expression  = "BOOL(1)"
  description = "my test description"

  dataset = "%s"
}`, alias, dataset),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("honeycombio_derived_column.test", "alias", alias),
					resource.TestCheckResourceAttr("honeycombio_derived_column.test", "expression", "BOOL(1)"),
					resource.TestCheckResourceAttr("honeycombio_derived_column.test", "description", "my test description"),
				),
			},
			{
				ResourceName:      "honeycombio_derived_column.test",
				ImportStateId:     fmt.Sprintf("%s/%s", dataset, alias),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "honeycombio_derived_column" "invalid_column_in_expression" {
  alias       = "%s"
  expression  = "LOG10($invalid_column)"

  dataset = "%s"
}`, test.RandomStringWithPrefix("test.", 10), dataset),
				ExpectError: regexp.MustCompile(`unknown column`),
			},
		},
	})
}
