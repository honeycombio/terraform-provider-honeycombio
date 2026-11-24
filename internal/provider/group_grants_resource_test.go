package provider

import (
	"context"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/test"
)

func TestAcc_GroupGrantsResource(t *testing.T) {
	envID, ds, slugs := testAccDatasetsForRestriction(t, 3)
	name := test.RandomStringWithPrefix("test.", 20)

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheckV2API(t),
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactory,
		CheckDestroy:             testAccEnsureGroupDestroyed(t),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigGroupGrants(name, envID, ds, ds[0], ds[1]),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("honeycombio_group_grants.test", "id", "honeycombio_group.test", "id"),
					resource.TestCheckResourceAttrPair("honeycombio_group_grants.test", "group_id", "honeycombio_group.test", "id"),
					resource.TestCheckResourceAttr("honeycombio_group_grants.test", "dataset_ids.#", "2"),
					testAccEnsureGroupGrantsAre(t, ds[0], ds[1]),
				),
			},
			{ // changing the Datasets is an update, not a replacement
				Config: testAccConfigGroupGrants(name, envID, ds, ds[1], ds[2]),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("honeycombio_group_grants.test", plancheck.ResourceActionUpdate),
					},
				},
				Check: testAccEnsureGroupGrantsAre(t, ds[1], ds[2]),
			},
			{
				ResourceName:      "honeycombio_group_grants.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{ // a grant made outside Terraform is revoked
				PreConfig: func() {
					err := testAccV2Client(t).Groups.GrantDatasets(context.Background(), testAccGroupIDByName(t, name), []string{ds[0]})
					require.NoError(t, err)
				},
				Config: testAccConfigGroupGrants(name, envID, ds, ds[1], ds[2]),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("honeycombio_group_grants.test", plancheck.ResourceActionUpdate),
					},
				},
				Check: testAccEnsureGroupGrantsAre(t, ds[1], ds[2]),
			},
			{ // a Dataset deleted outside Terraform drops out of the list
				PreConfig: func() {
					testAccDeleteDataset(t, slugs[2])
				},
				Config: testAccConfigGroupGrants(name, envID, ds[:2], ds[0], ds[1]),
				Check:  testAccEnsureGroupGrantsAre(t, ds[0], ds[1]),
			},
			{ // deleting the resource revokes every grant
				Config: testAccConfigGroupGrants(name, envID, ds[:2]),
				Check:  testAccEnsureGroupGrantsAre(t),
			},
		},
	})
}

// A Group with grants, such as grants made in the UI or by
// honeycombio_group_grant, must be imported.
func TestAcc_GroupGrantsResource_alreadyHasGrants(t *testing.T) {
	envID, ds, _ := testAccDatasetsForRestriction(t, 1)
	name := test.RandomStringWithPrefix("test.", 20)

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheckV2API(t),
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactory,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigGroupGrants(name, envID, ds),
			},
			{
				PreConfig: func() {
					err := testAccV2Client(t).Groups.GrantDatasets(context.Background(), testAccGroupIDByName(t, name), ds)
					require.NoError(t, err)
				},
				Config:      testAccConfigGroupGrants(name, envID, ds, ds[0]),
				ExpectError: regexp.MustCompile(`Honeycomb Group Already Has Grants`),
			},
			{ // the failed create left the grant in place
				Config: testAccConfigGroupGrants(name, envID, ds),
				Check:  testAccEnsureGroupGrantsAre(t, ds[0]),
			},
		},
	})
}

func TestAcc_GroupGrantsResource_errors(t *testing.T) {
	_, datasetID, _ := testAccDatasetForRestriction(t)
	name := test.RandomStringWithPrefix("test.", 20)

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheckV2API(t),
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactory,
		Steps: []resource.TestStep{
			{
				Config:      testAccConfigGroupGrantsTo(name, datasetID),
				ExpectError: regexp.MustCompile(`only a restricted dataset can be granted`),
			},
			{
				Config:      testAccConfigGroupGrantsTo(name, "hcxds_01aaaaaaaaaaaaaaaaaaaaaaaa"),
				ExpectError: regexp.MustCompile(`hcxds_01aaaaaaaaaaaaaaaaaaaaaaaa: no dataset with this ID`),
			},
			{
				Config: `
resource "honeycombio_group_grants" "test" {
  group_id    = "hcxgr_01aaaaaaaaaaaaaaaaaaaaaaaa"
  dataset_ids = []
}`,
				ExpectError: regexp.MustCompile(`The API found no Group "hcxgr_01aaaaaaaaaaaaaaaaaaaaaaaa"`),
			},
		},
	})
}

// testAccDatasetsForRestriction creates n datasets with
// testAccDatasetForRestriction, and returns their Environment's ID with
// the datasets' IDs and slugs.
func testAccDatasetsForRestriction(t *testing.T, n int) (environmentID string, datasetIDs, slugs []string) {
	t.Helper()

	for range n {
		envID, datasetID, slug := testAccDatasetForRestriction(t)
		environmentID = envID
		datasetIDs = append(datasetIDs, datasetID)
		slugs = append(slugs, slug)
	}
	return environmentID, datasetIDs, slugs
}

func testAccDeleteDataset(t *testing.T, slug string) {
	t.Helper()

	ctx := context.Background()
	c := testAccClient(t)
	_, err := c.Datasets.Update(ctx, &client.Dataset{
		Slug:     slug,
		Settings: client.DatasetSettings{DeleteProtected: new(false)},
	})
	require.NoError(t, err)
	require.NoError(t, c.Datasets.Delete(ctx, slug))
}

// testAccConfigGroupGrants is honeycombio_group.test, a
// honeycombio_dataset_restriction for each of restricted, and
// honeycombio_group_grants.test with granted. With no granted, the config
// has no honeycombio_group_grants.
func testAccConfigGroupGrants(groupName, environmentID string, restricted []string, granted ...string) string {
	var b strings.Builder
	b.WriteString(testAccConfigGroup(groupName, ""))
	for i, id := range restricted {
		fmt.Fprintf(&b, `
resource "honeycombio_dataset_restriction" "ds%[1]d" {
  environment_id = "%[2]s"
  dataset_id     = "%[3]s"
}
`, i, environmentID, id)
	}
	if len(granted) == 0 {
		return b.String()
	}

	refs := make([]string, len(granted))
	for i, id := range granted {
		refs[i] = fmt.Sprintf("honeycombio_dataset_restriction.ds%d.dataset_id", slices.Index(restricted, id))
	}
	fmt.Fprintf(&b, `
resource "honeycombio_group_grants" "test" {
  group_id    = honeycombio_group.test.id
  dataset_ids = [%s]
}`, strings.Join(refs, ", "))
	return b.String()
}

// testAccConfigGroupGrantsTo grants honeycombio_group.test the Dataset
// datasetID, with no restriction in the config.
func testAccConfigGroupGrantsTo(groupName, datasetID string) string {
	return testAccConfigGroup(groupName, "") + fmt.Sprintf(`
resource "honeycombio_group_grants" "test" {
  group_id    = honeycombio_group.test.id
  dataset_ids = ["%s"]
}`, datasetID)
}

// testAccEnsureGroupGrantsAre checks that the API lists exactly datasetIDs
// as the grants of honeycombio_group.test.
func testAccEnsureGroupGrantsAre(t *testing.T, datasetIDs ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		got, err := testAccGroupGrantIDs(t, s, datasetTarget)
		if err != nil {
			return err
		}
		want := slices.Clone(datasetIDs)
		slices.Sort(got)
		slices.Sort(want)
		if !slices.Equal(got, want) {
			return fmt.Errorf("Group grants are %v in the API, want %v", got, want)
		}
		return nil
	}
}
