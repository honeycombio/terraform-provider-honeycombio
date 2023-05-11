package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/honeycombio/terraform-provider-honeycombio/client"
)

func TestAcc_DerivedColumnDataSource(t *testing.T) {
	ctx := context.Background()
	c := testAccClient(t)
	dataset := testAccDataset()

	testColumn := client.DerivedColumn{
		Alias:       acctest.RandString(4) + "_column1",
		Description: "test column1",
		Expression:  "BOOL(1)",
	}

	col, err := c.DerivedColumns.Create(ctx, dataset, &testColumn)
	if err != nil {
		t.Error(err)
	}
	// update ID for removal later
	testColumn.ID = col.ID
	t.Cleanup(func() {
		//nolint:errcheck
		c.DerivedColumns.Delete(ctx, dataset, testColumn.ID)
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
		Steps: []resource.TestStep{
			{
				Config: testAccDerivedColumnDataSourceConfig(dataset, testColumn.Alias),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.honeycombio_derived_column.test", "description", testColumn.Description),
					resource.TestCheckResourceAttr("data.honeycombio_derived_column.test", "expression", testColumn.Expression),
				),
			},
		},
	})
}

func testAccDerivedColumnDataSourceConfig(dataset, alias string) string {
	return fmt.Sprintf(`
data "honeycombio_derived_column" "test" {
  dataset     = "%s"
  alias       = "%s"
}
`, dataset, alias)
}
