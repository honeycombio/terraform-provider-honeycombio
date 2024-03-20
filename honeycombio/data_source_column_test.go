package honeycombio

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"

	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
)

func TestAccDataSourceHoneycombioColumn_basic(t *testing.T) {
	ctx := context.Background()
	c := testAccClient(t)
	dataset := testAccDataset()

	testColumns := []honeycombio.Column{
		{
			KeyName:     acctest.RandString(4) + "_test_column3",
			Description: "test column3",
			Type:        honeycombio.ToPtr(honeycombio.ColumnType("float")),
		},
	}

	for i, column := range testColumns {
		col, err := c.Columns.Create(ctx, dataset, &column)
		require.NoError(t, err)
		// update ID for removal later
		testColumns[i].ID = col.ID

	}
	//nolint:errcheck
	t.Cleanup(func() {
		// remove Columns at the of the test run
		for _, col := range testColumns {
			c.Columns.Delete(ctx, dataset, col.ID)
		}
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
		Steps: []resource.TestStep{
			// match by name and return a single column with the right type
			{
				Config: testAccDataSourceColumnConfig([]string{"dataset = \"" + testAccDataset() + "\"", "name = \"" + testColumns[0].KeyName + "\""}),
				Check:  resource.TestCheckResourceAttr("data.honeycombio_column.test", "type", "float"),
			},
			// test a failed match
			{
				Config:      testAccDataSourceColumnConfig([]string{"dataset = \"" + testAccDataset() + "\"", "name = \"test_column5\""}),
				ExpectError: regexp.MustCompile("404 Not Found"),
			},
		},
	})
}

func testAccDataSourceColumnConfig(filters []string) string {
	return fmt.Sprintf(`
data "honeycombio_column" "test" {
	%s
}

output "type" {
  value = data.honeycombio_column.test.type
}

output "description" {
  value = data.honeycombio_column.test.description
}
`, strings.Join(filters, "\n"))
}
