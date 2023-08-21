package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
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
	if err != nil {
		t.Error(err)
	}
	slo, err := c.SLOs.Create(ctx, dataset, &client.SLO{
		Name:             acctest.RandString(4) + "_slo",
		Description:      "test SLO",
		TimePeriodDays:   30,
		TargetPerMillion: 995000,
		SLI:              client.SLIRef{Alias: sli.Alias},
	})
	if err != nil {
		t.Error(err)
	}

	//nolint:errcheck
	t.Cleanup(func() {
		c.SLOs.Delete(ctx, dataset, slo.ID)
		c.DerivedColumns.Delete(ctx, dataset, sli.ID)
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
		Steps: []resource.TestStep{
			{
				Config: testAccSLODataSourceConfig(dataset, slo.ID),
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

func testAccSLODataSourceConfig(dataset, id string) string {
	return fmt.Sprintf(`
data "honeycombio_slo" "test" {
  id      = "%s"
  dataset = "%s"
}
`, id, dataset)
}
