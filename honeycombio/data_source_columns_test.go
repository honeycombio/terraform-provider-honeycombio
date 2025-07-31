package honeycombio

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/test"
)

func TestAccDataSourceHoneycombioColumns_basic(t *testing.T) {
	ctx := context.Background()
	c := testAccClient(t)
	dataset := testAccDataset()

	const numColumns = 5
	testFilterPrefix := test.RandomStringWithPrefix("test.", 5)
	testColumns := make([]*honeycombio.Column, 0, numColumns)
	for range numColumns {
		col, err := c.Columns.Create(ctx, dataset, &honeycombio.Column{
			KeyName:     test.RandomStringWithPrefix(testFilterPrefix+".", 10),
			Description: test.RandomString(20),
			Type:        honeycombio.ToPtr(honeycombio.ColumnTypeFloat),
		})
		require.NoError(t, err)
		testColumns = append(testColumns, col)
	}
	t.Cleanup(func() {
		for _, col := range testColumns {
			_ = c.Columns.Delete(ctx, dataset, col.ID)
		}
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "honeycombio_columns" "all" {
  dataset = "%[1]s"
}

data "honeycombio_columns" "filtered" {
  dataset     = "%[1]s"
  starts_with = "%[2]s"
}

data "honeycombio_columns" "none" {
  dataset     = "%[1]s"
  starts_with = "does-not-exist"
}

output "all" {
  value = data.honeycombio_columns.all.names
}`, dataset, testFilterPrefix),
				Check: resource.ComposeAggregateTestCheckFunc(
					testCheckAllOutputContains(testColumns[0].KeyName),
					testCheckAllOutputContains(testColumns[1].KeyName),
					testCheckAllOutputContains(testColumns[2].KeyName),
					testCheckAllOutputContains(testColumns[3].KeyName),
					testCheckAllOutputContains(testColumns[4].KeyName),
					resource.TestCheckResourceAttr("data.honeycombio_columns.filtered",
						"names.#",
						fmt.Sprintf("%d", numColumns),
					),
					resource.TestCheckResourceAttr("data.honeycombio_columns.none",
						"names.#",
						"0",
					),
				),
				PlanOnly: true,
			},
		},
	})
}

func testCheckAllOutputContains(contains string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		const name = "all"

		ms := s.RootModule()
		rs, ok := ms.Outputs[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		output, ok := rs.Value.([]interface{})
		if !ok {
			return fmt.Errorf("output value is not a list")
		}

		for _, value := range output {
			valueStr, ok := value.(string)
			if !ok {
				return fmt.Errorf("output value is not a string")
			}
			if valueStr == contains {
				return nil
			}
		}

		return fmt.Errorf("Output '%s' did not contain %#v, got %#v", name, contains, output)
	}
}
