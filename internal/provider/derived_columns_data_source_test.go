package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/test"
)

func TestAcc_DerivedColumnsDataSource(t *testing.T) {
	ctx := t.Context()
	c := testAccClient(t)

	t.Run("dataset-specific lookup", func(t *testing.T) {
		dataset := testAccDataset()
		dcPrefix := fmt.Sprintf("test.%s-", test.RandomString(8))

		testColumns := []client.DerivedColumn{
			{
				Alias:       dcPrefix + "column1",
				Description: "test column1",
				Expression:  "BOOL(1)",
			},
			{
				Alias:       dcPrefix + "column2",
				Description: "test column2",
				Expression:  "BOOL(1)",
			},
		}

		for i, column := range testColumns {
			col, err := c.DerivedColumns.Create(ctx, dataset, &column)
			require.NoError(t, err)

			// update ID for removal later
			testColumns[i].ID = col.ID
		}
		t.Cleanup(func() {
			// remove DCs at the of the test run
			for _, col := range testColumns {
				c.DerivedColumns.Delete(ctx, dataset, col.ID)
			}
		})

		resource.Test(t, resource.TestCase{
			PreCheck:                 testAccPreCheck(t),
			ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
			Steps: []resource.TestStep{
				{
					Config: fmt.Sprintf(`
data "honeycombio_derived_columns" "test" {
  dataset     = "%s"
  starts_with = "%s"
}`, dataset, dcPrefix),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.honeycombio_derived_columns.test", "names.#", fmt.Sprintf("%d", len(testColumns))),
					),
				},
			},
		})
	})

	t.Run("environment-wide lookup", func(t *testing.T) {
		if c.IsClassic(ctx) {
			t.Skip("classic does not support environment-wide derived columns")
		}

		dcPrefix := fmt.Sprintf("test.%s-", test.RandomString(8))
		testColumns := []client.DerivedColumn{
			{
				Alias:       dcPrefix + "column1",
				Description: "test column1",
				Expression:  "BOOL(1)",
			},
			{
				Alias:       dcPrefix + "column2",
				Description: "test column2",
				Expression:  "BOOL(1)",
			},
		}

		for i, column := range testColumns {
			col, err := c.DerivedColumns.Create(ctx, client.EnvironmentWideSlug, &column)
			require.NoError(t, err)

			// update ID for removal later
			testColumns[i].ID = col.ID
		}
		t.Cleanup(func() {
			// remove DCs at the of the test run
			for _, col := range testColumns {
				c.DerivedColumns.Delete(ctx, client.EnvironmentWideSlug, col.ID)
			}
		})

		resource.Test(t, resource.TestCase{
			PreCheck:                 testAccPreCheck(t),
			ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
			Steps: []resource.TestStep{
				{
					Config: fmt.Sprintf(`
data "honeycombio_derived_columns" "test" {
  starts_with = "%s"
}`, dcPrefix),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.honeycombio_derived_columns.test", "names.#", fmt.Sprintf("%d", len(testColumns))),
					),
				},
			},
		})
	})
}
