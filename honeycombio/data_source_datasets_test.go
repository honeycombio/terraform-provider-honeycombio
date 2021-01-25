package honeycombio

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceHoneycombioDataset_basic(t *testing.T) {
	dataset := testAccDataset()

	resource.Test(t, resource.TestCase{
		PreCheck:          testAccPreCheck(t),
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceDatasetConfig(nil),
				Check: resource.ComposeTestCheckFunc(
					testCheckOutputContains("names", dataset),
				),
			},
			{
				Config: testAccDataSourceDatasetConfig([]string{"starts_with = \"kvrhdn/terraform-\""}),
				Check: resource.ComposeTestCheckFunc(
					testCheckOutputContains("names", dataset),
				),
			},
			{
				Config: testAccDataSourceDatasetConfig([]string{"starts_with = \"foo\""}),
				Check: resource.ComposeTestCheckFunc(
					testCheckOutputDoesNotContain("names", dataset),
				),
			},
		},
	})
}

func testAccDataSourceDatasetConfig(filters []string) string {
	return fmt.Sprintf(`
data "honeycombio_datasets" "test" {
	%s
}

output "names" {
  value = data.honeycombio_datasets.test.names
}`, strings.Join(filters, "\n"))
}
