package provider

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"

	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/test"
)

func TestAcc_GroupGrantResource(t *testing.T) {
	envID, datasetID, _ := testAccDatasetForRestriction(t)
	name := test.RandomStringWithPrefix("test.", 20)
	config := testAccConfigGroupGrant(name, envID, datasetID)

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheckV2API(t),
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactory,
		CheckDestroy:             testAccEnsureGroupDestroyed(t),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("honeycombio_group_grant.test", "group_id", "honeycombio_group.test", "id"),
					resource.TestCheckResourceAttr("honeycombio_group_grant.test", "dataset_id", datasetID),
					resource.TestCheckResourceAttrWith("honeycombio_group_grant.test", "id", func(id string) error {
						if want := testAccGroupIDByName(t, name) + "/dataset/" + datasetID; id != want {
							return fmt.Errorf("id is %q, want %q", id, want)
						}
						return nil
					}),
					testAccEnsureGroupGrant(t, datasetTarget, datasetID, true),
				),
			},
			{
				ResourceName:      "honeycombio_group_grant.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{ // a grant revoked outside Terraform is created again
				PreConfig: func() {
					err := testAccV2Client(t).Groups.RevokeDatasets(context.Background(), testAccGroupIDByName(t, name), []string{datasetID})
					require.NoError(t, err)
				},
				Config: config,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("honeycombio_group_grant.test", plancheck.ResourceActionCreate),
					},
				},
				Check: testAccEnsureGroupGrant(t, datasetTarget, datasetID, true),
			},
			{ // deleting the grant revokes it, and the Dataset stays restricted
				Config: testAccConfigGroupAndRestriction(name, envID, datasetID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccEnsureGroupGrant(t, datasetTarget, datasetID, false),
					testAccEnsureDatasetRestricted(t, envID, datasetID, true),
				),
			},
		},
	})
}

func TestAcc_GroupGrantResource_environment(t *testing.T) {
	envID := testAccEnvironmentID(t)
	name := test.RandomStringWithPrefix("test.", 20)
	config := testAccConfigGroup(name, "") + `
resource "honeycombio_group_grant" "test" {
  group_id       = honeycombio_group.test.id
  environment_id = "` + envID + `"
}`

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheckV2API(t),
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactory,
		CheckDestroy:             testAccEnsureGroupDestroyed(t),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("honeycombio_group_grant.test", "environment_id", envID),
					resource.TestCheckNoResourceAttr("honeycombio_group_grant.test", "dataset_id"),
					resource.TestCheckResourceAttrWith("honeycombio_group_grant.test", "id", func(id string) error {
						if want := testAccGroupIDByName(t, name) + "/environment/" + envID; id != want {
							return fmt.Errorf("id is %q, want %q", id, want)
						}
						return nil
					}),
					testAccEnsureGroupGrant(t, environmentTarget, envID, true),
				),
			},
			{
				ResourceName:      "honeycombio_group_grant.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{ // a grant revoked outside Terraform is created again
				PreConfig: func() {
					err := testAccV2Client(t).Groups.RevokeEnvironments(context.Background(), testAccGroupIDByName(t, name), []string{envID})
					require.NoError(t, err)
				},
				Config: config,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("honeycombio_group_grant.test", plancheck.ResourceActionCreate),
					},
				},
				Check: testAccEnsureGroupGrant(t, environmentTarget, envID, true),
			},
			{ // deleting the grant revokes it
				Config: testAccConfigGroup(name, ""),
				Check:  testAccEnsureGroupGrant(t, environmentTarget, envID, false),
			},
			{ // a grant made outside Terraform must be imported
				PreConfig: func() {
					err := testAccV2Client(t).Groups.GrantEnvironments(context.Background(), testAccGroupIDByName(t, name), []string{envID})
					require.NoError(t, err)
				},
				Config:      config,
				ExpectError: regexp.MustCompile(`Honeycomb Group Grant Already Exists`),
			},
			{
				Config: testAccConfigGroup(name, "") + `
resource "honeycombio_group_grant" "test" {
  group_id       = honeycombio_group.test.id
  environment_id = "hcxen_01aaaaaaaaaaaaaaaaaaaaaaaa"
}`,
				ExpectError: regexp.MustCompile(`hcxen_01aaaaaaaaaaaaaaaaaaaaaaaa: no environment with this ID`),
			},
		},
	})
}

func TestGroupGrantResource_needsExactlyOneTarget(t *testing.T) {
	// the validator fails the plan before the provider calls the API
	for name, tc := range map[string]struct{ targets, wantErr string }{
		"neither": {``, `Missing Attribute Configuration`},
		"both": {
			`dataset_id = "hcxds_01aaaaaaaaaaaaaaaaaaaaaaaa"` + "\n" + `environment_id = "hcxen_01aaaaaaaaaaaaaaaaaaaaaaaa"`,
			`Invalid Attribute Combination`,
		},
	} {
		t.Run(name, func(t *testing.T) {
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactory,
				Steps: []resource.TestStep{
					{
						Config: `
resource "honeycombio_group_grant" "test" {
  group_id = "hcxgr_01aaaaaaaaaaaaaaaaaaaaaaaa"
  ` + tc.targets + `
}`,
						PlanOnly:    true,
						ExpectError: regexp.MustCompile(tc.wantErr),
					},
				},
			})
		})
	}
}

// A grant made outside Terraform, such as in the UI, must be imported, so
// that deleting the resource can't revoke a grant it didn't create.
func TestAcc_GroupGrantResource_alreadyExists(t *testing.T) {
	envID, datasetID, _ := testAccDatasetForRestriction(t)
	name := test.RandomStringWithPrefix("test.", 20)

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheckV2API(t),
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactory,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigGroupAndRestriction(name, envID, datasetID),
			},
			{
				PreConfig: func() {
					err := testAccV2Client(t).Groups.GrantDatasets(context.Background(), testAccGroupIDByName(t, name), []string{datasetID})
					require.NoError(t, err)
				},
				Config:      testAccConfigGroupGrant(name, envID, datasetID),
				ExpectError: regexp.MustCompile(`Honeycomb Group Grant Already Exists`),
			},
			{ // the failed create left the grant in place
				Config: testAccConfigGroupAndRestriction(name, envID, datasetID),
				Check:  testAccEnsureGroupGrant(t, datasetTarget, datasetID, true),
			},
		},
	})
}

// The API deletes a Dataset's grants when the Dataset loses its restriction.
func TestAcc_GroupGrantResource_datasetUnrestricted(t *testing.T) {
	envID, datasetID, _ := testAccDatasetForRestriction(t)
	name := test.RandomStringWithPrefix("test.", 20)

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheckV2API(t),
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactory,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigGroupGrant(name, envID, datasetID),
			},
			{ // the grant and the restriction leave state, so the plan creates both again
				PreConfig: func() {
					testAccSetDatasetRestricted(t, envID, datasetID, false)
				},
				RefreshState:       true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAcc_GroupGrantResource_errors(t *testing.T) {
	_, datasetID, _ := testAccDatasetForRestriction(t)
	name := test.RandomStringWithPrefix("test.", 20)
	config := testAccConfigGroupGrantTo(name, `"`+datasetID+`"`)

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheckV2API(t),
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactory,
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile(`only a restricted dataset can be granted`),
			},
			{
				Config:      testAccConfigGroupGrantTo(name, `"hcxds_01aaaaaaaaaaaaaaaaaaaaaaaa"`),
				ExpectError: regexp.MustCompile(`hcxds_01aaaaaaaaaaaaaaaaaaaaaaaa: no dataset with this ID`),
			},
			{ // a Group that wasn't granted the Dataset has no grant to import
				Config:            config,
				ResourceName:      "honeycombio_group_grant.test",
				ImportState:       true,
				ImportStateIdFunc: testAccGroupGrantImportID(datasetID, "/dataset/"),
				ExpectError:       regexp.MustCompile(`Cannot import non-existent remote object`),
			},
			{
				Config:            config,
				ResourceName:      "honeycombio_group_grant.test",
				ImportState:       true,
				ImportStateIdFunc: testAccGroupGrantImportID(datasetID, "/pattern/"),
				ExpectError:       regexp.MustCompile(`Import ID must be written as <group_id>/dataset/<dataset_id> or`),
			},
			{
				Config: testAccConfigGroup(name, "") + `
resource "honeycombio_group_grant" "test" {
  group_id   = "hcxgr_01aaaaaaaaaaaaaaaaaaaaaaaa"
  dataset_id = "` + datasetID + `"
}`,
				ExpectError: regexp.MustCompile(`The API found no Group "hcxgr_01aaaaaaaaaaaaaaaaaaaaaaaa"`),
			},
		},
	})
}

// testAccGroupGrantImportID joins the ID of honeycombio_group.test in state
// and datasetID with sep.
func testAccGroupGrantImportID(datasetID, sep string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources["honeycombio_group.test"]
		if !ok {
			return "", fmt.Errorf("resource not found: honeycombio_group.test")
		}
		return rs.Primary.ID + sep + datasetID, nil
	}
}

// testAccConfigGroupAndRestriction is a Group and a restricted Dataset,
// with no grant.
func testAccConfigGroupAndRestriction(groupName, environmentID, datasetID string) string {
	return testAccConfigGroup(groupName, "") + testAccConfigDatasetRestriction(environmentID, datasetID)
}

func testAccConfigGroupGrant(groupName, environmentID, datasetID string) string {
	return testAccConfigGroupAndRestriction(groupName, environmentID, datasetID) + `
resource "honeycombio_group_grant" "test" {
  group_id   = honeycombio_group.test.id
  dataset_id = honeycombio_dataset_restriction.test.dataset_id
}`
}

// testAccConfigGroupGrantTo grants honeycombio_group.test the Dataset that
// the HCL expression datasetID names.
func testAccConfigGroupGrantTo(groupName, datasetID string) string {
	return testAccConfigGroup(groupName, "") + `
resource "honeycombio_group_grant" "test" {
  group_id   = honeycombio_group.test.id
  dataset_id = ` + datasetID + `
}`
}

// testAccEnsureGroupGrant checks whether the API lists id as a grant of
// the kind, datasetTarget or environmentTarget, for honeycombio_group.test.
func testAccEnsureGroupGrant(t *testing.T, kind, id string, want bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ids, err := testAccGroupGrantIDs(t, s, kind)
		if err != nil {
			return err
		}
		if got := slices.Contains(ids, id); got != want {
			return fmt.Errorf("grant of %s %s is %t in the API, want %t", kind, id, got, want)
		}
		return nil
	}
}

// testAccGroupGrantIDs returns the IDs that the API lists as grants of
// the kind for honeycombio_group.test.
func testAccGroupGrantIDs(t *testing.T, s *terraform.State, kind string) ([]string, error) {
	rs, ok := s.RootModule().Resources["honeycombio_group.test"]
	if !ok {
		return nil, fmt.Errorf("resource not found: honeycombio_group.test")
	}
	for _, k := range grantKinds(testAccV2Client(t)) {
		if k.name == kind {
			return k.list(context.Background(), rs.Primary.ID)
		}
	}
	return nil, fmt.Errorf("unknown grant kind %q", kind)
}

// TestGroupGrantResource_fakeAPI checks that the resource asks the API for its
// own grant only. An older API ignores the filter and lists every grant, and the
// resource must work with that API too.
func TestGroupGrantResource_fakeAPI(t *testing.T) {
	const (
		groupID   = "hcxgr_01jbr2v3k8m4n5p6q7r8s9t0vw"
		datasetID = "hcxds_01jbr4k2m7n8p9q0r1s2t3v4wx"
		otherID   = "hcxds_01jbr4k2m7n8p9q0r1s2t3v4xy"
		listPath  = "/2/teams/fake/groups/" + groupID + "/relationships/datasets"
	)

	for _, api := range []struct {
		name        string
		filterWorks bool
	}{
		{"API with the filter", true},
		{"API without the filter", false},
	} {
		t.Run(api.name, func(t *testing.T) {
			var mu sync.Mutex
			granted := []string{otherID}
			var listQueries []string

			mux := http.NewServeMux()
			mux.HandleFunc("GET /2/auth", func(w http.ResponseWriter, _ *http.Request) {
				testWriteJSONAPI(w, http.StatusOK, `{
					"data": {"id": "k", "type": "api-keys",
						"attributes": {"name": "k", "key_type": "management", "scopes": ["groups:write"]},
						"relationships": {"team": {"data": {"type": "teams", "id": "t"}}}},
					"included": [{"id": "t", "type": "teams", "attributes": {"name": "Fake", "slug": "fake"}}]}`)
			})
			mux.HandleFunc("GET "+listPath, func(w http.ResponseWriter, r *http.Request) {
				mu.Lock()
				defer mu.Unlock()
				want := r.URL.Query().Get("filter[id]")
				listQueries = append(listQueries, want)
				if api.filterWorks && want != "" && !strings.HasPrefix(want, "hcxds_") {
					testWriteJSONAPI(w, http.StatusUnprocessableEntity, `{"errors": [{"status": "422",
						"title": "invalid filter", "detail": "not a Dataset ID", "source": {"parameter": "filter[id]"}}]}`)
					return
				}
				var items []string
				for _, id := range granted {
					if !api.filterWorks || want == "" || want == id {
						items = append(items, fmt.Sprintf(`{"type": "datasets", "id": "%s"}`, id))
					}
				}
				testWriteJSONAPI(w, http.StatusOK, `{"data": [`+strings.Join(items, ",")+`], "links": {"next": null}}`)
			})
			mux.HandleFunc("POST "+listPath, func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				if !strings.Contains(string(body), datasetID) {
					testWriteJSONAPI(w, http.StatusUnprocessableEntity, `{"errors": [{"status": "422",
						"title": "invalid dataset", "detail": "no such Dataset", "source": {"pointer": "/data/0/id"}}]}`)
					return
				}
				mu.Lock()
				defer mu.Unlock()
				granted = append(granted, datasetID)
				w.WriteHeader(http.StatusNoContent)
			})
			mux.HandleFunc("DELETE "+listPath, func(w http.ResponseWriter, _ *http.Request) {
				mu.Lock()
				defer mu.Unlock()
				granted = slices.DeleteFunc(granted, func(id string) bool { return id == datasetID })
				w.WriteHeader(http.StatusNoContent)
			})
			srv := httptest.NewServer(mux)
			t.Cleanup(srv.Close)

			t.Setenv("HONEYCOMB_API_ENDPOINT", srv.URL)
			t.Setenv("HONEYCOMB_API_KEY", "")
			t.Setenv("HONEYCOMB_KEY_ID", "fake")
			t.Setenv("HONEYCOMB_KEY_SECRET", "fake")

			configFor := func(id string) string {
				return `
resource "honeycombio_group_grant" "test" {
  group_id   = "` + groupID + `"
  dataset_id = "` + id + `"
}`
			}
			everyListFiltered := func(*terraform.State) error {
				mu.Lock()
				defer mu.Unlock()
				if len(listQueries) == 0 {
					return errors.New("the resource sent no list request")
				}
				for _, q := range listQueries {
					if q != datasetID {
						return fmt.Errorf("a list request had filter[id]=%q, want %q", q, datasetID)
					}
				}
				return nil
			}
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactory,
				Steps: []resource.TestStep{
					{
						Config: configFor(datasetID),
						Check:  everyListFiltered,
					},
					{ // a grant revoked outside Terraform is found missing, even when another grant is listed
						PreConfig: func() {
							mu.Lock()
							defer mu.Unlock()
							granted = []string{otherID}
						},
						Config: configFor(datasetID),
						ConfigPlanChecks: resource.ConfigPlanChecks{
							PreApply: []plancheck.PlanCheck{
								plancheck.ExpectResourceAction("honeycombio_group_grant.test", plancheck.ResourceActionCreate),
							},
						},
						Check: everyListFiltered,
					},
					{ // a slug is not a Dataset ID: the filter rejects it, and the grant's own error is shown
						Config:      configFor("my-dataset"),
						ExpectError: regexp.MustCompile(`no such Dataset`),
					},
				},
			})
		})
	}
}
