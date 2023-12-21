package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
)

func TestAcc_DerivedColumnsDataSource(t *testing.T) {
	ctx := context.Background()
	c := testAccClient(t)
	dataset := testAccDataset()
	dcPrefix := acctest.RandString(4)

	testColumns := []client.DerivedColumn{
		{
			Alias:       dcPrefix + "_column1",
			Description: "test column1",
			Expression:  "BOOL(1)",
		},
		{
			Alias:       dcPrefix + "_column2",
			Description: "test column2",
			Expression:  "BOOL(1)",
		},
	}

	for i, column := range testColumns {
		col, err := c.DerivedColumns.Create(ctx, dataset, &column)
		if err != nil {
			t.Error(err)
		}
		// update ID for removal later
		testColumns[i].ID = col.ID
	}
	t.Cleanup(func() {
		// remove DCs at the of the test run
		for _, col := range testColumns {
			//nolint:errcheck
			c.DerivedColumns.Delete(ctx, dataset, col.ID)
		}
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
		Steps: []resource.TestStep{
			{
				Config: testAccDerivedColumnsDataSourceConfig(dataset, dcPrefix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.honeycombio_derived_columns.test", "names.#", fmt.Sprintf("%d", len(testColumns))),
				),
			},
		},
	})
}

func testAccDerivedColumnsDataSourceConfig(dataset, prefix string) string {
	return fmt.Sprintf(`
data "honeycombio_derived_columns" "test" {
  dataset     = "%s"
  starts_with = "%s"
}
`, dataset, prefix)
}
