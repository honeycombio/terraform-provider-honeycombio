package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/test"
)

func TestAcc_DerivedColumnDataSource(t *testing.T) {
	ctx := context.Background()
	c := testAccClient(t)

	t.Run("dataset-specific lookup", func(t *testing.T) {
		dataset := testAccDataset()

		testColumn := client.DerivedColumn{
			Alias:       test.RandomStringWithPrefix("test.", 10),
			Description: test.RandomString(20),
			Expression:  "BOOL(1)",
		}

		col, err := c.DerivedColumns.Create(ctx, dataset, &testColumn)
		require.NoError(t, err)

		// update ID for removal later
		testColumn.ID = col.ID
		t.Cleanup(func() {
			c.DerivedColumns.Delete(ctx, dataset, testColumn.ID)
		})

		resource.Test(t, resource.TestCase{
			PreCheck:                 testAccPreCheck(t),
			ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
			Steps: []resource.TestStep{
				{
					Config: fmt.Sprintf(`
data "honeycombio_derived_column" "test" {
  dataset     = "%s"
  alias       = "%s"
}`, dataset, testColumn.Alias),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.honeycombio_derived_column.test", "description", testColumn.Description),
						resource.TestCheckResourceAttr("data.honeycombio_derived_column.test", "expression", testColumn.Expression),
					),
				},
			},
		})
	})

	t.Run("environment-wide lookup", func(t *testing.T) {
		if c.IsClassic(ctx) {
			t.Skip("classic does not support environment-wide derived columns")
		}

		testColumn := client.DerivedColumn{
			Alias:       test.RandomStringWithPrefix("test.", 10),
			Description: test.RandomString(20),
			Expression:  "BOOL(1)",
		}

		col, err := c.DerivedColumns.Create(ctx, client.EnvironmentWideSlug, &testColumn)
		require.NoError(t, err)

		// update ID for removal later
		testColumn.ID = col.ID
		t.Cleanup(func() {
			c.DerivedColumns.Delete(ctx, client.EnvironmentWideSlug, testColumn.ID)
		})

		resource.Test(t, resource.TestCase{
			PreCheck:                 testAccPreCheck(t),
			ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
			Steps: []resource.TestStep{
				{
					Config: fmt.Sprintf(`
data "honeycombio_derived_column" "test" {
  alias       = "%s"
}`, testColumn.Alias),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.honeycombio_derived_column.test", "description", testColumn.Description),
						resource.TestCheckResourceAttr("data.honeycombio_derived_column.test", "expression", testColumn.Expression),
					),
				},
			},
		})
	})
}
