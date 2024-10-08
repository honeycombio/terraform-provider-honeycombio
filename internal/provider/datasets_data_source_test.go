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

func TestAcc_DatasetsDatasource(t *testing.T) {
	ctx := context.Background()
	c := testAccClient(t)
	const numDatasets = 15

	// create a bunch of datasets
	testDatasets := make([]*client.Dataset, numDatasets)
	for i := range numDatasets {
		d, err := c.Datasets.Create(ctx, &client.Dataset{
			Name:        test.RandomStringWithPrefix("test.ds.", 20),
			Description: test.RandomString(70),
		})
		require.NoError(t, err)
		testDatasets[i] = d
	}

	t.Cleanup(func() {
		for _, d := range testDatasets {
			// disable deletion protection and delete the Dataset
			c.Datasets.Update(ctx, &client.Dataset{
				Slug: d.Slug,
				Settings: client.DatasetSettings{
					DeleteProtected: helper.ToPtr(false),
				},
			})
			c.Datasets.Delete(ctx, d.Slug)
		}
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheckV2API(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "honeycombio_datasets" "all" {}

data "honeycombio_datasets" "regex" {
  detail_filter {
    name        = "name"
    value_regex = "test.ds.*"
  }
}

data "honeycombio_datasets" "starts_with" {
  starts_with = "test.ds."
}

data "honeycombio_datasets" "exact" {
  detail_filter {
    name  = "name"
    value = "%s"
  }
}`, testDatasets[0].Name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.honeycombio_datasets.regex", "slugs.#", fmt.Sprintf("%d", numDatasets)),
					resource.TestCheckResourceAttr("data.honeycombio_datasets.starts_with", "slugs.#", fmt.Sprintf("%d", numDatasets)),
					resource.TestCheckResourceAttr("data.honeycombio_datasets.exact", "slugs.#", "1"),
				),
			},
		},
	})
}
