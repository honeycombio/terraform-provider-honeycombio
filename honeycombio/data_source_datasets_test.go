package honeycombio

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
)

func TestAccDataSourceHoneycombioDataset_basic(t *testing.T) {
	createArgs := honeycombio.DatasetCreateArgs{
		Description:     "",
		ExpandJSONDepth: 0,
	}

	dataset := testAccDatasetWithArgs(createArgs)

	resource.Test(t, resource.TestCase{
		PreCheck:          testAccPreCheck(t),
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceDatasetConfig(nil),
				Check: resource.ComposeTestCheckFunc(
					testCheckOutputContains("names", dataset.Name),
				),
			},
			{
				Config: testAccDataSourceDatasetConfig([]string{"starts_with = \"" + string(dataset.Name[0:2]) + "\""}),
				Check: resource.ComposeTestCheckFunc(
					testCheckOutputContains("names", dataset.Name),
				),
			},
			{
				Config: testAccDataSourceDatasetConfig([]string{"starts_with = \"foo\""}),
				Check: resource.ComposeTestCheckFunc(
					testCheckOutputDoesNotContain("names", dataset.Name),
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
