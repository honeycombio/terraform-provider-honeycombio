package provider

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"

	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/test"
)

func TestAcc_ColumnResource(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		dataset := testAccDataset()
		name := test.RandomStringWithPrefix("test.", 10)

		resource.Test(t, resource.TestCase{
			PreCheck:                 testAccPreCheck(t),
			ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
			Steps: []resource.TestStep{
				{
					Config: fmt.Sprintf(`
resource "honeycombio_column" "test" {
  name        = "%s"
  type        = "float"
  hidden      = false
  description = "Duration of the trace"

  dataset = "%s"
}`, name, dataset),
					Check: resource.ComposeTestCheckFunc(
						testAccEnsureColumnExists(t, "honeycombio_column.test", name),
						resource.TestCheckResourceAttrSet("honeycombio_column.test", "id"),
						resource.TestCheckResourceAttr("honeycombio_column.test", "name", name),
						resource.TestCheckResourceAttr("honeycombio_column.test", "dataset", dataset),
						resource.TestCheckResourceAttr("honeycombio_column.test", "type", "float"),
						resource.TestCheckResourceAttr("honeycombio_column.test", "hidden", "false"),
						resource.TestCheckResourceAttr("honeycombio_column.test", "description", "Duration of the trace"),
						resource.TestCheckResourceAttrSet("honeycombio_column.test", "created_at"),
						resource.TestCheckResourceAttrSet("honeycombio_column.test", "updated_at"),
						resource.TestCheckResourceAttrSet("honeycombio_column.test", "last_written_at"),
					),
				},
				{
					Config: fmt.Sprintf(`
resource "honeycombio_column" "test" {
  name        = "%s"
  dataset     = "%s"
  type        = "float"
  hidden      = true
  description = "My nice column"
}`, name, dataset),
					RefreshState: false, // skip refresh here prevent racey updates
				},
				{
					// updating columns can be racey so we wait a bit to ensure the update has propagated
					// to the API before checking the state.
					PreConfig: func() {
						client := testAccClient(t)

						require.Eventually(t, func() bool {
							column, err := client.Columns.GetByKeyName(context.Background(), dataset, name)
							return err == nil && column.Description == "My nice column"
						}, 15*time.Second, 100*time.Millisecond, "Column update did not complete in time")
					},
					Config: fmt.Sprintf(`
resource "honeycombio_column" "test" {
  name        = "%s"
  dataset     = "%s"
  type        = "float"
  hidden      = true
  description = "My nice column"
}`, name, dataset),
					ExpectNonEmptyPlan: false,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("honeycombio_column.test", "id"),
						resource.TestCheckResourceAttr("honeycombio_column.test", "name", name),
						resource.TestCheckResourceAttr("honeycombio_column.test", "dataset", dataset),
						resource.TestCheckResourceAttr("honeycombio_column.test", "type", "float"),
						resource.TestCheckResourceAttr("honeycombio_column.test", "hidden", "true"),
						resource.TestCheckResourceAttr("honeycombio_column.test", "description", "My nice column"),
						resource.TestCheckResourceAttrSet("honeycombio_column.test", "created_at"),
						resource.TestCheckResourceAttrSet("honeycombio_column.test", "updated_at"),
						resource.TestCheckResourceAttrSet("honeycombio_column.test", "last_written_at"),
					),
				},
				{
					ResourceName:      "honeycombio_column.test",
					ImportStateId:     fmt.Sprintf("%s/%s", dataset, name),
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})
}

// TestAcc_ColumnResourceUpgradeFromVersion037 is intended to test the migration
// case from the last SDK-based version of the Column resource to the current Framework-based
// version.
//
// See: https://developer.hashicorp.com/terraform/plugin/framework/migrating/testing#testing-migration
func TestAcc_ColumnResourceUpgradeFromVersion037(t *testing.T) {
	dataset := testAccDataset()
	name := test.RandomStringWithPrefix("test.", 10)

	config := fmt.Sprintf(`
resource "honeycombio_column" "test" {
  name        = "%s"
  dataset     = "%s"
  description = "My nice column"
}`, name, dataset)

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"honeycombio": {
						VersionConstraint: "0.37.1",
						Source:            "honeycombio/honeycombio",
					},
				},
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccEnsureColumnExists(t, "honeycombio_column.test", name),
				),
			},
			{
				ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
				Config:                   config,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func testAccEnsureColumnExists(t *testing.T, resource, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resourceState, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("%q not found in state", resource)
		}

		client := testAccClient(t)
		_, err := client.Columns.GetByKeyName(context.Background(), resourceState.Primary.Attributes["dataset"], name)
		if err != nil {
			return fmt.Errorf("failed to fetch created column: %w", err)
		}

		return nil
	}
}
