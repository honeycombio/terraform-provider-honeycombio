package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/test"
)

func TestAcc_DatsetDataSource(t *testing.T) {
	ctx := context.Background()
	c := testAccClient(t)

	ds, err := c.Datasets.Create(ctx, &client.Dataset{
		Name:            test.RandomStringWithPrefix("test.", 20),
		Description:     test.RandomString(70),
		ExpandJSONDepth: 3,
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		// disable deletion protection and delete the Dataset
		c.Datasets.Update(ctx, &client.Dataset{
			Slug: ds.Slug,
			Settings: client.DatasetSettings{
				DeleteProtected: helper.ToPtr(false),
			},
		})
		c.Datasets.Delete(ctx, ds.Slug)
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheckV2API(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "honeycombio_dataset" "test" {
  slug = "%s"
}`, ds.Slug),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.honeycombio_dataset.test", "id", ds.Slug),
					resource.TestCheckResourceAttr("data.honeycombio_dataset.test", "name", ds.Name),
					resource.TestCheckResourceAttr("data.honeycombio_dataset.test", "slug", ds.Slug),
					resource.TestCheckResourceAttr("data.honeycombio_dataset.test", "expand_json_depth", "3"),
					resource.TestCheckResourceAttr("data.honeycombio_dataset.test", "description", ds.Description),
					resource.TestCheckResourceAttr("data.honeycombio_dataset.test", "delete_protected", "true"),
				),
			},
		},
	})
}
