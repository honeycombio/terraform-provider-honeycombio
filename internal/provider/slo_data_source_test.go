package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper"
)

func TestAcc_SLODataSource(t *testing.T) {
	ctx := context.Background()
	c := testAccClient(t)
	dataset := testAccDataset()

	sli, err := c.DerivedColumns.Create(ctx, dataset, &client.DerivedColumn{
		Alias:       acctest.RandString(4) + "_sli",
		Description: "test SLI",
		Expression:  "BOOL(1)",
	})
	require.NoError(t, err)

	slo, err := c.SLOs.Create(ctx, dataset, &client.SLO{
		Name:             acctest.RandString(4) + "_slo",
		Description:      "test SLO",
		TimePeriodDays:   30,
		TargetPerMillion: 995000,
		SLI:              client.SLIRef{Alias: sli.Alias},
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		c.SLOs.Delete(ctx, dataset, slo.ID)
		c.DerivedColumns.Delete(ctx, dataset, sli.ID)
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "honeycombio_slo" "test" {
  id      = "%s"
  dataset = "%s"
}`, slo.ID, dataset),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.honeycombio_slo.test", "name", slo.Name),
					resource.TestCheckResourceAttr("data.honeycombio_slo.test", "description", slo.Description),
					resource.TestCheckResourceAttr("data.honeycombio_slo.test", "sli", slo.SLI.Alias),
					resource.TestCheckResourceAttr("data.honeycombio_slo.test", "target_percentage", "99.5"),
					resource.TestCheckResourceAttr("data.honeycombio_slo.test", "time_period", "30"),
				),
			},
		},
	})
}

func TestAcc_MDSLODataSource(t *testing.T) {
	ctx := context.Background()
	c := testAccClient(t)

	if c.IsClassic(ctx) {
		t.Skip("Classic does not support multi-dataset SLOs")
	}

	dataset_all := "__all__"
	dataset1, err := c.Datasets.Create(ctx, &client.Dataset{
		Name:        "dataset1",
		Description: "test dataset 1",
	})
	require.NoError(t, err)

	dataset2, err := c.Datasets.Create(ctx, &client.Dataset{
		Name:        "dataset2",
		Description: "test dataset 2",
	})
	require.NoError(t, err)

	sli, err := c.DerivedColumns.Create(ctx, dataset_all, &client.DerivedColumn{
		Alias:       acctest.RandString(4) + "_sli",
		Description: "test SLI",
		Expression:  "BOOL(1)",
	})
	require.NoError(t, err)

	slo, err := c.SLOs.Create(ctx, dataset_all, &client.SLO{
		Name:             acctest.RandString(4) + "_slo",
		Description:      "test SLO",
		TimePeriodDays:   30,
		TargetPerMillion: 995000,
		SLI:              client.SLIRef{Alias: sli.Alias},
		DatasetSlugs:     []string{dataset1.Slug, dataset2.Slug},
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		c.SLOs.Delete(ctx, dataset_all, slo.ID)
		c.DerivedColumns.Delete(ctx, dataset_all, sli.ID)

		c.Datasets.Update(ctx, &client.Dataset{
			Slug: dataset1.Slug,
			Settings: client.DatasetSettings{
				DeleteProtected: helper.ToPtr(false),
			},
		})
		err = c.Datasets.Delete(ctx, dataset1.Slug)

		c.Datasets.Update(ctx, &client.Dataset{
			Slug: dataset2.Slug,
			Settings: client.DatasetSettings{
				DeleteProtected: helper.ToPtr(false),
			},
		})
		err = c.Datasets.Delete(ctx, dataset2.Slug)
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "honeycombio_slo" "test" {
  id      = "%s"
}`, slo.ID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.honeycombio_slo.test", "name", slo.Name),
					resource.TestCheckResourceAttr("data.honeycombio_slo.test", "description", slo.Description),
					resource.TestCheckResourceAttr("data.honeycombio_slo.test", "sli", slo.SLI.Alias),
					resource.TestCheckResourceAttr("data.honeycombio_slo.test", "target_percentage", "99.5"),
					resource.TestCheckResourceAttr("data.honeycombio_slo.test", "time_period", "30"),
					resource.TestCheckResourceAttr("data.honeycombio_slo.test", "dataset_slugs.#", "2"),
					resource.TestCheckResourceAttr("data.honeycombio_slo.test", "dataset_slugs.0", dataset1.Slug),
					resource.TestCheckResourceAttr("data.honeycombio_slo.test", "dataset_slugs.1", dataset2.Slug),
				),
			},
		},
	})
}
