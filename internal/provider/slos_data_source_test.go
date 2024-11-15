package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
)

func TestAcc_SLOsDataSource(t *testing.T) {
	ctx := context.Background()
	c := testAccClient(t)
	dataset := testAccDataset()
	testPrefix := acctest.RandString(8)

	testData := []struct {
		SLI client.DerivedColumn
		SLO client.SLO
	}{
		{
			SLI: client.DerivedColumn{
				Alias:      testPrefix + "_sli1",
				Expression: "BOOL(1)",
			},
			SLO: client.SLO{
				Name:             testPrefix + "_slo1",
				SLI:              client.SLIRef{Alias: testPrefix + "_sli1"},
				TimePeriodDays:   30,
				TargetPerMillion: 995000,
			},
		},
		{
			SLI: client.DerivedColumn{
				Alias:      testPrefix + "_sli2",
				Expression: "BOOL(1)",
			},
			SLO: client.SLO{
				Name:             testPrefix + "_slo2",
				SLI:              client.SLIRef{Alias: testPrefix + "_sli2"},
				TimePeriodDays:   30,
				TargetPerMillion: 995000,
			},
		},
		{
			SLI: client.DerivedColumn{
				Alias:      testPrefix + "_sli3",
				Expression: "BOOL(1)",
			},
			SLO: client.SLO{
				// different prefix for all vs filtered testing
				Name:             acctest.RandString(8) + "_slo",
				SLI:              client.SLIRef{Alias: testPrefix + "_sli3"},
				TimePeriodDays:   30,
				TargetPerMillion: 995000,
			},
		},
	}

	for i, tc := range testData {
		sli, err := c.DerivedColumns.Create(ctx, dataset, &tc.SLI)
		require.NoError(t, err)
		slo, err := c.SLOs.Create(ctx, dataset, &tc.SLO)
		require.NoError(t, err)

		// update IDs for removal later
		testData[i].SLI.ID = sli.ID
		testData[i].SLO.ID = slo.ID
	}

	t.Cleanup(func() {
		// remove SLOs at the of the test run
		for _, tc := range testData {
			c.SLOs.Delete(ctx, dataset, tc.SLO.ID)
			c.DerivedColumns.Delete(ctx, dataset, tc.SLI.ID)
		}
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "honeycombio_slos" "all" {
  dataset = "%[1]s"
}

data "honeycombio_slos" "regex" {
  dataset     = "%[1]s"

  detail_filter {
    name        = "name"
    value_regex = "%[2]s*"
  }
}

data "honeycombio_slos" "exact" {
  dataset     = "%[1]s"

  detail_filter {
    name  = "name"
    value = "%[2]s_slo1"
  }
}
`, dataset, testPrefix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.honeycombio_slos.regex", "ids.#", "2"),
					resource.TestCheckResourceAttr("data.honeycombio_slos.exact", "ids.#", "1"),
				),
			},
		},
	})
}
