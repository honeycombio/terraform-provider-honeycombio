package honeycombio

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"

	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
)

func TestAccDataSourceHoneycombioColumns_basic(t *testing.T) {
	ctx := context.Background()
	c := testAccClient(t)
	dataset := testAccDataset()

	testprefix := acctest.RandString(4)

	testColumns := []honeycombio.Column{
		{
			KeyName:     testprefix + "_test_column1",
			Description: "test column1",
		},
		{
			KeyName:     testprefix + "_test_column2",
			Description: "test column2",
		},
	}

	for i, column := range testColumns {
		col, err := c.Columns.Create(ctx, dataset, &column)
		// update ID for removal later
		testColumns[i].ID = col.ID
		if err != nil {
			t.Error(err)
		}
	}
	t.Cleanup(func() {
		// remove Columns at the of the test run
		for _, col := range testColumns {
			c.Columns.Delete(ctx, dataset, col.ID)
		}
	})

	resource.Test(t, resource.TestCase{
		PreCheck:          testAccPreCheck(t),
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceColumnsConfig([]string{"dataset = \"" + testAccDataset() + "\""}),
				Check: resource.ComposeTestCheckFunc(
					testCheckOutputContains("names", testColumns[0].KeyName),
				),
			},
			{
				Config: testAccDataSourceColumnsConfig([]string{"dataset = \"" + testAccDataset() + "\"", "starts_with = \"" + testprefix + "\""}),
				Check:  resource.TestCheckResourceAttr("data.honeycombio_columns.test", "names.#", "2"),
			},
			{
				Config: testAccDataSourceColumnsConfig([]string{"dataset = \"" + testAccDataset() + "\"", "starts_with = \"foo\""}),
				Check: resource.ComposeTestCheckFunc(
					testCheckOutputDoesNotContain("names", testColumns[0].KeyName),
				),
			},
		},
	})
}

func testAccDataSourceColumnsConfig(filters []string) string {
	return fmt.Sprintf(`
data "honeycombio_columns" "test" {
	%s
}

output "names" {
  value = data.honeycombio_columns.test.names
}`, strings.Join(filters, "\n"))
}
