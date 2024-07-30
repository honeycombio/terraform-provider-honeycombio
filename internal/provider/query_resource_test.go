package provider

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/test"
)

func TestAcc_QueryResource(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	dataset := testAccDataset()
	c := testAccClient(t)
	col, err := c.Columns.Create(ctx, dataset, &client.Column{
		KeyName: test.RandomStringWithPrefix("test.", 10),
		Type:    client.ToPtr(client.ColumnTypeFloat),
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		c.Columns.Delete(ctx, dataset, col.ID)
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigBasicQueryTest(dataset, col.KeyName, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccEnsureQueryExists(t, "honeycombio_query.test"),
					resource.TestCheckResourceAttrSet("honeycombio_query.test", "id"),
					resource.TestCheckResourceAttrSet("honeycombio_query.test", "query_json"),
					resource.TestCheckResourceAttr("honeycombio_query.test", "dataset", dataset),
				),
			},
			{
				Config: testAccConfigBasicQueryTest(dataset, col.KeyName, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccEnsureQueryExists(t, "honeycombio_query.test"),
					resource.TestCheckResourceAttrSet("honeycombio_query.test", "id"),
					resource.TestCheckResourceAttrSet("honeycombio_query.test", "query_json"),
					resource.TestCheckResourceAttr("honeycombio_query.test", "dataset", dataset),
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				ResourceName:        "honeycombio_query.test",
				ImportStateIdPrefix: fmt.Sprintf("%v/", dataset),
				ImportState:         true,
			},
		},
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
		Steps: []resource.TestStep{
			{
				Config: `
resource "honeycombio_query" "test" {
  dataset = "does-not-matter"
  query_json = <<EOT
"invalid": "json"]
EOT
}`,
				ExpectError: regexp.MustCompile(`json: cannot unmarshal string`),
			},
			{
				Config: `
resource "honeycombio_query" "test" {
  dataset = "does-not-matter"
  query_json = <<EOT
{
  "calculations": [
    {
      "op": "COUNT"
    }
  ],
  "query_window": 7200
}
EOT
}`,
				ExpectError: regexp.MustCompile(`json: unknown field "query_window"`),
			},
		},
	})
}

// TestAcc_QueryResourceUpgradeFromVersion022 is intended to test the migration
// case from the last SDK-based version of the Query resource to the current Framework-based
// version.
//
// See: https://developer.hashicorp.com/terraform/plugin/framework/migrating/testing#testing-migration
func TestAcc_QueryResourceUpgradeFromVersion022(t *testing.T) {
	t.Skip("mysteriously broken and under investigation")

	t.Parallel()
	ctx := context.Background()
	dataset := testAccDataset()
	c := testAccClient(t)
	col, err := c.Columns.Create(ctx, dataset, &client.Column{
		KeyName: test.RandomStringWithPrefix("test.", 10),
		Type:    client.ToPtr(client.ColumnTypeFloat),
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		c.Columns.Delete(ctx, dataset, col.ID)
	})
	config := testAccConfigBasicQueryTest(dataset, col.KeyName, 1)

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"honeycombio": {
						VersionConstraint: "~> 0.22.0",
						Source:            "honeycombio/honeycombio",
					},
				},
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccEnsureQueryExists(t, "honeycombio_query.test"),
				),
			},
			{
				ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
				Config:                   config,
				PlanOnly:                 true,
				ExpectNonEmptyPlan:       false,
			},
		},
	})
}

// TestAcc_QueryResourceEquivalentQuerySpecSupressed tests the  behavior of the
// resource when an equivalent query is suppressed by the plan modifier.
func TestAcc_QueryResourceEquivalentQuerySpecSupressed(t *testing.T) {
	t.Parallel()
	dataset := testAccDataset()

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "honeycombio_query" "test" {
  dataset    = "%s"
  query_json = "{}"
}`, dataset),
			},
			{
				// The query is technically equivalent to the previous one, so it should be suppressed.
				//  n.b. spec copied from the UI from a simple 'COUNT' query
				Config: fmt.Sprintf(`
resource "honeycombio_query" "test" {
  dataset    = "%s"
  query_json = <<EOT
{
  "time_range": 7200,
  "granularity": 0,
  "breakdowns": [],
  "calculations": [
    {
      "op": "COUNT"
    }
  ],
  "orders": [],
  "havings": [],
  "limit": 1000
}
EOT
}`, dataset),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccConfigBasicQueryTest(dataset, column string, value float64) string {
	return fmt.Sprintf(`
data "honeycombio_query_specification" "test" {
  calculation {
    op     = "AVG"
    column = "%[2]s"
  }

  filter {
    column = "%[2]s"
    op     = ">"
    value  = %[3]f
  }
}

resource "honeycombio_query" "test" {
  dataset    = "%[1]s"
  query_json = data.honeycombio_query_specification.test.json
}`, dataset, column, value)
}

func testAccEnsureQueryExists(t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resourceState, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("\"%s\" not found in state", name)
		}

		client := testAccClient(t)
		_, err := client.Queries.Get(context.Background(), resourceState.Primary.Attributes["dataset"], resourceState.Primary.ID)
		if err != nil {
			return fmt.Errorf("failed to fetch created query: %w", err)
		}

		return nil
	}
}
