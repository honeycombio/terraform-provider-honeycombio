package honeycombio

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/test"
)

func TestAccDataSourceHoneycombioColumn_basic(t *testing.T) {
	ctx := context.Background()
	c := testAccClient(t)
	dataset := testAccDataset()

	col, err := c.Columns.Create(ctx, dataset, &honeycombio.Column{
		KeyName:     test.RandomStringWithPrefix("test.", 10),
		Description: test.RandomString(20),
		Type:        honeycombio.ToPtr(honeycombio.ColumnTypeFloat),
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = c.Columns.Delete(ctx, dataset, col.ID)
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "honeycombio_column" "test" {
  dataset = "%s"
  name    = "%s"
}`, dataset, col.KeyName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.honeycombio_column.test", "name", col.KeyName),
					resource.TestCheckResourceAttr("data.honeycombio_column.test", "description", col.Description),
					resource.TestCheckResourceAttr("data.honeycombio_column.test", "type", "float"),
					resource.TestCheckResourceAttr("data.honeycombio_column.test", "hidden", "false"),
					resource.TestCheckResourceAttrSet("data.honeycombio_column.test", "last_written_at"),
					resource.TestCheckResourceAttrSet("data.honeycombio_column.test", "created_at"),
					resource.TestCheckResourceAttrSet("data.honeycombio_column.test", "updated_at"),
				),
				PlanOnly: true,
			},
		},
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "honeycombio_column" "test" {
  dataset = "%s"
  name    = "does-not-exist"
}`, dataset),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`(?i)not found`),
			},
		},
	})
}
