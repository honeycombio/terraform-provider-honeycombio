package provider

import (
	"context"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/test"
)

func TestAcc_GroupGrantsDataSource(t *testing.T) {
	envID, datasetID, _ := testAccDatasetForRestriction(t)
	_, outsideID, _ := testAccDatasetForRestriction(t)
	testAccSetDatasetRestricted(t, envID, outsideID, true)
	name := test.RandomStringWithPrefix("test.", 20)

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheckV2API(t),
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactory,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigGroup(name, "") + testAccConfigGroupGrantsDataSource(""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.honeycombio_group_grants.test", "id", "honeycombio_group.test", "id"),
					resource.TestCheckResourceAttr("data.honeycombio_group_grants.test", "dataset_ids.#", "0"),
				),
			},
			{ // lists the grants that Terraform made and the grants made outside Terraform
				PreConfig: func() {
					err := testAccV2Client(t).Groups.GrantDatasets(context.Background(), testAccGroupIDByName(t, name), []string{outsideID})
					require.NoError(t, err)
				},
				Config: testAccConfigGroupGrant(name, envID, datasetID) +
					testAccConfigGroupGrantsDataSource("depends_on = [honeycombio_group_grant.test]"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.honeycombio_group_grants.test", "dataset_ids.#", "2"),
					resource.TestCheckTypeSetElemAttr("data.honeycombio_group_grants.test", "dataset_ids.*", datasetID),
					resource.TestCheckTypeSetElemAttr("data.honeycombio_group_grants.test", "dataset_ids.*", outsideID),
				),
			},
			{
				Config: `
data "honeycombio_group_grants" "test" {
  group_id = "hcxgr_01aaaaaaaaaaaaaaaaaaaaaaaa"
}`,
				ExpectError: regexp.MustCompile(`The API found no Group "hcxgr_01aaaaaaaaaaaaaaaaaaaaaaaa"`),
			},
		},
	})
}

// testAccConfigGroupGrantsDataSource reads the grants of
// honeycombio_group.test. extra is added to the data block.
func testAccConfigGroupGrantsDataSource(extra string) string {
	return `
data "honeycombio_group_grants" "test" {
  group_id = honeycombio_group.test.id
  ` + extra + `
}`
}
