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

# Multiple filters
data "honeycombio_slos" "multi_filter" {
  dataset     = "%[1]s"

  detail_filter {
    name    = "name"
    operator = "starts-with"
    value    = "%[2]s"
  }
  
  detail_filter {
    name    = "name"
    operator = "ends-with"
    value    = "slo1"
  }
}

# Multiple filters with different names
data "honeycombio_slos" "different_names" {
  dataset     = "%[1]s"

  detail_filter {
    name    = "name"
    operator = "starts-with"
    value    = "%[2]s"
  }
  
  detail_filter {
    name    = "time_period"
    operator = "equals"
    value    = "30"
  }
}

`, dataset, testPrefix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.honeycombio_slos.regex", "ids.#", "2"),
					resource.TestCheckResourceAttr("data.honeycombio_slos.exact", "ids.#", "1"),
					resource.TestCheckResourceAttr("data.honeycombio_slos.multi_filter", "ids.#", "1"),
					resource.TestCheckResourceAttr("data.honeycombio_slos.different_names", "ids.#", "2"),
				),
			},
		},
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "honeycombio_slos" "regex" {

  detail_filter {
    name        = "name"
    value_regex = "%[1]s*"
  }
}
`, testPrefix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.honeycombio_slos.regex", "ids.#", "2"),
				),
			},
		},
	})

	// Test for case without dataset specified
	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "honeycombio_slos" "regex" {
  detail_filter {
    name        = "name"
    value_regex = "%[1]s*"
  }
}

# Test combined multiple operators
data "honeycombio_slos" "combined_operators" {
  detail_filter {
    name    = "name"
    operator = "contains"
    value    = "%[1]s"
  }
  
  detail_filter {
    name    = "name"
    operator = "not-equals"
    value    = "%[1]s_slo3"  # This doesn't exist but ensures NOT logic works
  }
}
`, testPrefix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.honeycombio_slos.regex", "ids.#", "2"),
					resource.TestCheckResourceAttr("data.honeycombio_slos.combined_operators", "ids.#", "2"),
				),
			},
		},
	})
}

// Test specifically for filter groups
func TestAcc_SLOsDataSource_FilterGroups(t *testing.T) {
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
				Alias:      testPrefix + "_sli_high",
				Expression: "BOOL(1)",
			},
			SLO: client.SLO{
				Name:             testPrefix + "_high_slo",
				SLI:              client.SLIRef{Alias: testPrefix + "_sli_high"},
				TimePeriodDays:   30,
				TargetPerMillion: 999000,
				Description:      "High reliability SLO",
			},
		},
		{
			SLI: client.DerivedColumn{
				Alias:      testPrefix + "_sli_medium",
				Expression: "BOOL(1)",
			},
			SLO: client.SLO{
				Name:             testPrefix + "_medium_slo",
				SLI:              client.SLIRef{Alias: testPrefix + "_sli_medium"},
				TimePeriodDays:   7,
				TargetPerMillion: 995000,
				Description:      "Medium reliability SLO",
			},
		},
		{
			SLI: client.DerivedColumn{
				Alias:      testPrefix + "_sli_low",
				Expression: "BOOL(1)",
			},
			SLO: client.SLO{
				Name:             testPrefix + "_low_slo",
				SLI:              client.SLIRef{Alias: testPrefix + "_sli_low"},
				TimePeriodDays:   1,
				TargetPerMillion: 990000,
				Description:      "Low reliability SLO",
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
# Test combining different name filters
data "honeycombio_slos" "combined_names" {
  dataset = "%[1]s"

  detail_filter {
    name    = "name"
    operator = "contains"
    value    = "%[2]s"
  }
  
  detail_filter {
    name    = "description"
    operator = "contains"
    value    = "reliability"
  }
}

# Test numeric range filtering (medium target range)
data "honeycombio_slos" "numeric_range" {
  dataset = "%[1]s"

  detail_filter {
    name    = "target_percentage"
    operator = ">"
    value    = "99.1"
  }
  
  detail_filter {
    name    = "target_percentage"
    operator = "<"
    value    = "99.7"
  }
}

# Test combining different operators
data "honeycombio_slos" "complex_filter" {
  dataset = "%[1]s"

  detail_filter {
    name    = "name"
    operator = "starts-with"
    value    = "%[2]s"
  }
  
  detail_filter {
    name    = "name"
    operator = "not-equals"
    value    = "%[2]s_low_slo"
  }
  
  detail_filter {
    name    = "time_period"
    operator = ">"
    value    = "1"
  }
}
`, dataset, testPrefix),
				Check: resource.ComposeAggregateTestCheckFunc(
					// All SLOs should match the combined names filter
					resource.TestCheckResourceAttr("data.honeycombio_slos.combined_names", "ids.#", "3"),

					// Only medium target SLO should be in this range
					resource.TestCheckResourceAttr("data.honeycombio_slos.numeric_range", "ids.#", "1"),

					// High and medium SLOs should match complex filter
					resource.TestCheckResourceAttr("data.honeycombio_slos.complex_filter", "ids.#", "2"),
				),
			},
		},
	})
}

// Test for SLO filtering by tags
func TestAcc_SLOsDataSource_TagFiltering(t *testing.T) {
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
				Alias:      testPrefix + "_sli_prod",
				Expression: "BOOL(1)",
			},
			SLO: client.SLO{
				Name:             testPrefix + "_slo_prod",
				SLI:              client.SLIRef{Alias: testPrefix + "_sli_prod"},
				TimePeriodDays:   30,
				TargetPerMillion: 999000,
				Tags: []client.Tag{
					{Key: "env", Value: "production"},
					{Key: "level", Value: "critical"},
					{Key: "team", Value: "core"},
				},
			},
		},
		{
			SLI: client.DerivedColumn{
				Alias:      testPrefix + "_sli_staging",
				Expression: "BOOL(1)",
			},
			SLO: client.SLO{
				Name:             testPrefix + "_slo_staging",
				SLI:              client.SLIRef{Alias: testPrefix + "_sli_staging"},
				TimePeriodDays:   30,
				TargetPerMillion: 995000,
				Tags: []client.Tag{
					{Key: "env", Value: "staging"},
					{Key: "team", Value: "core"},
				},
			},
		},
		{
			SLI: client.DerivedColumn{
				Alias:      testPrefix + "_sli_dev",
				Expression: "BOOL(1)",
			},
			SLO: client.SLO{
				Name:             testPrefix + "_slo_dev",
				SLI:              client.SLIRef{Alias: testPrefix + "_sli_dev"},
				TimePeriodDays:   30,
				TargetPerMillion: 990000,
				Tags: []client.Tag{
					{Key: "env", Value: "development"},
					{Key: "team", Value: "platform"},
				},
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
# Filter SLOs by production tag
data "honeycombio_slos" "production_slos" {
  dataset = "%[1]s"

  detail_filter {
    name     = "tags"
    operator = "contains"
    value    = "production"
  }
}

# Filter SLOs by team:core tag
data "honeycombio_slos" "core_team_slos" {
  dataset = "%[1]s"

  detail_filter {
    name     = "tags"
    operator = "contains"
    value    = "team:core"
  }
}

# Filter SLOs by team:platform tag
data "honeycombio_slos" "platform_team_slos" {
  dataset = "%[1]s"

  detail_filter {
    name     = "tags"
    operator = "contains"
    value    = "team:platform"
  }
}

# Filter SLOs that don't have the development tag
data "honeycombio_slos" "non_dev_slos" {
  dataset = "%[1]s"

  detail_filter {
    name     = "tags"
    operator = "does-not-contain"
    value    = "development"
  }
}

# Filter SLOs with combined tag and name filters
data "honeycombio_slos" "core_team_with_staging" {
  dataset = "%[1]s"

  detail_filter {
    name     = "tags"
    operator = "contains"
    value    = "team:core"
  }
  
  detail_filter {
    name     = "name"
    operator = "contains"
    value    = "staging"
  }
}

# Filter with regex on tags
data "honeycombio_slos" "team_regex_slos" {
  dataset = "%[1]s"

  detail_filter {
    name        = "tags"
    value_regex = "team:.*"
  }
}
`, dataset, testPrefix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.honeycombio_slos.production_slos", "ids.#", "1"),
					resource.TestCheckResourceAttr("data.honeycombio_slos.core_team_slos", "ids.#", "2"),
					resource.TestCheckResourceAttr("data.honeycombio_slos.platform_team_slos", "ids.#", "1"),
					resource.TestCheckResourceAttr("data.honeycombio_slos.non_dev_slos", "ids.#", "2"),
					resource.TestCheckResourceAttr("data.honeycombio_slos.core_team_with_staging", "ids.#", "1"),
					resource.TestCheckResourceAttr("data.honeycombio_slos.team_regex_slos", "ids.#", "3"),
				),
			},
		},
	})
}
