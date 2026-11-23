package provider

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
	v2client "github.com/honeycombio/terraform-provider-honeycombio/client/v2"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/test"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/test/fixture"
)

func TestAcc_DatasetRestrictionResource(t *testing.T) {
	envID, datasetID, _ := testAccDatasetForRestriction(t)
	config := testAccConfigDatasetRestriction(envID, datasetID)

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheckV2API(t),
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactory,
		// the dataset stays, without its restriction
		CheckDestroy: testAccEnsureDatasetRestricted(t, envID, datasetID, false),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("honeycombio_dataset_restriction.test", "id", envID+"/"+datasetID),
					resource.TestCheckResourceAttr("honeycombio_dataset_restriction.test", "environment_id", envID),
					resource.TestCheckResourceAttr("honeycombio_dataset_restriction.test", "dataset_id", datasetID),
					testAccEnsureDatasetRestricted(t, envID, datasetID, true),
				),
			},
			{
				ResourceName:      "honeycombio_dataset_restriction.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{ // a restriction removed outside Terraform is created again
				PreConfig: func() {
					testAccSetDatasetRestricted(t, envID, datasetID, false)
				},
				Config: config,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("honeycombio_dataset_restriction.test", plancheck.ResourceActionCreate),
					},
				},
				Check: testAccEnsureDatasetRestricted(t, envID, datasetID, true),
			},
		},
	})
}

// A dataset restricted outside Terraform, such as in the UI, must be imported,
// so that deleting the resource can't remove a restriction it didn't create.
func TestAcc_DatasetRestrictionResource_alreadyRestricted(t *testing.T) {
	envID, datasetID, _ := testAccDatasetForRestriction(t)
	testAccSetDatasetRestricted(t, envID, datasetID, true)

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheckV2API(t),
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactory,
		CheckDestroy:             testAccEnsureDatasetRestricted(t, envID, datasetID, true),
		Steps: []resource.TestStep{
			{
				Config:      testAccConfigDatasetRestriction(envID, datasetID),
				ExpectError: regexp.MustCompile(`Honeycomb Dataset Is Already Restricted`),
			},
		},
	})
}

func TestAcc_DatasetRestrictionResource_datasetDeleted(t *testing.T) {
	envID, datasetID, slug := testAccDatasetForRestriction(t)
	config := testAccConfigDatasetRestriction(envID, datasetID)

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheckV2API(t),
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactory,
		Steps: []resource.TestStep{
			{
				Config: config,
			},
			{ // the restriction of a deleted dataset leaves state, so the plan creates it again
				PreConfig: func() {
					testAccDeleteDataset(t, slug)
				},
				RefreshState:       true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// TestDatasetRestrictionResource_fakeAPI covers the API's 403 for a key that
// can't revoke a dataset from its Groups. An acceptance test would need a
// second Management Key.
func TestDatasetRestrictionResource_fakeAPI(t *testing.T) {
	const (
		envID       = "hcxen_01j1jrsewaha3m0z6fwffpcrxg"
		datasetID   = "hcxds_01jbr4k2m7n8p9q0r1s2t3v4wx"
		datasetPath = "/2/teams/fake/environments/" + envID + "/datasets/" + datasetID
	)
	isRestricted := false
	writeDataset := func(w http.ResponseWriter) {
		testWriteJSONAPI(w, http.StatusOK, fmt.Sprintf(
			`{"data": {"id": "%s", "type": "datasets", "attributes": {"restricted": %t}}}`, datasetID, isRestricted))
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /2/auth", func(w http.ResponseWriter, _ *http.Request) {
		testWriteJSONAPI(w, http.StatusOK, `{
			"data": {"id": "k", "type": "api-keys",
				"attributes": {"name": "k", "key_type": "management", "scopes": ["datasets:write"]},
				"relationships": {"team": {"data": {"type": "teams", "id": "t"}}}},
			"included": [{"id": "t", "type": "teams", "attributes": {"name": "Fake", "slug": "fake"}}]}`)
	})
	mux.HandleFunc("GET "+datasetPath, func(w http.ResponseWriter, _ *http.Request) {
		writeDataset(w)
	})
	refused := false
	mux.HandleFunc("PATCH "+datasetPath, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		// only the first removal is refused, so the test's own destroy succeeds
		if strings.Contains(string(body), `"restricted":false`) && !refused {
			refused = true
			// the API's response to a key that can't revoke the dataset from a Group that has access to it
			testWriteJSONAPI(w, http.StatusForbidden, `{"errors": [{"id": "d5c966612468262cb8e4048e41dce194",
				"title": "this API key is not allowed to access this resource", "code": "permission-denied/not-allowed"}]}`)
			return
		}
		isRestricted = strings.Contains(string(body), `"restricted":true`)
		writeDataset(w)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	t.Setenv("HONEYCOMB_API_ENDPOINT", srv.URL)
	t.Setenv("HONEYCOMB_API_KEY", "")
	t.Setenv("HONEYCOMB_KEY_ID", "fake")
	t.Setenv("HONEYCOMB_KEY_SECRET", "fake")

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactory,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigDatasetRestriction(envID, datasetID),
			},
			{
				Config:      testAccConfigDatasetRestriction(envID, datasetID),
				Destroy:     true,
				ExpectError: regexp.MustCompile(`(?s)not allowed to access this resource.*cannot remove this Dataset Restriction`),
			},
		},
	})
}

func TestAcc_DatasetRestrictionResource_errors(t *testing.T) {
	envID, datasetID, _ := testAccDatasetForRestriction(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheckV2API(t),
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactory,
		Steps: []resource.TestStep{
			{ // a dataset that is not restricted has no restriction to import
				Config:        testAccConfigDatasetRestriction(envID, datasetID),
				ResourceName:  "honeycombio_dataset_restriction.test",
				ImportState:   true,
				ImportStateId: envID + "/" + datasetID,
				ExpectError:   regexp.MustCompile(`Cannot import non-existent remote object`),
			},
			{
				Config:        testAccConfigDatasetRestriction(envID, datasetID),
				ResourceName:  "honeycombio_dataset_restriction.test",
				ImportState:   true,
				ImportStateId: datasetID,
				ExpectError:   regexp.MustCompile(`Import ID must be written as <environment_id>/<dataset_id>`),
			},
			{ // a dataset is found only in its own Environment
				Config:      testAccConfigDatasetRestriction("hcxen_01aaaaaaaaaaaaaaaaaaaaaaaa", datasetID),
				ExpectError: regexp.MustCompile(`The API found no Dataset "` + datasetID + `" in\s+the\s+Environment`),
			},
		},
	})
}

func testAccConfigDatasetRestriction(environmentID, datasetID string) string {
	return fmt.Sprintf(`
resource "honeycombio_dataset_restriction" "test" {
  environment_id = "%s"
  dataset_id     = "%s"
}`, environmentID, datasetID)
}

// testAccDatasetForRestriction creates a dataset, and returns the ID of its
// Environment with the dataset's ID and slug. The dataset is deleted when
// the test completes.
func testAccDatasetForRestriction(t *testing.T) (environmentID, datasetID, slug string) {
	t.Helper()

	ctx := context.Background()
	c := testAccClient(t)
	v2 := testAccV2Client(t)

	ds := fixture.CreateDataset(ctx, t, c, &client.Dataset{
		Name: test.RandomStringWithPrefix("test.", 20),
	})

	environmentID = testAccEnvironmentID(t)

	datasets, err := v2.Datasets.List(ctx, environmentID, v2client.Filter("slug", ds.Slug))
	require.NoError(t, err)
	page, err := datasets.Next(ctx)
	require.NoError(t, err)
	require.Len(t, page, 1)

	return environmentID, page[0].ID, ds.Slug
}

// testAccEnvironmentID returns the ID of the provider API key's Environment.
func testAccEnvironmentID(t *testing.T) string {
	t.Helper()

	auth, err := testAccClient(t).Auth.List(context.Background())
	require.NoError(t, err)
	require.NotNil(t, auth.Environment.ID, "the API returned no Environment ID")
	return *auth.Environment.ID
}

func testAccSetDatasetRestricted(t *testing.T, environmentID, datasetID string, restricted bool) {
	t.Helper()

	_, err := testAccV2Client(t).Datasets.Update(context.Background(), environmentID, &v2client.Dataset{
		ID:         datasetID,
		Restricted: &restricted,
	})
	require.NoError(t, err)
}

func testAccEnsureDatasetRestricted(t *testing.T, environmentID, datasetID string, want bool) resource.TestCheckFunc {
	return func(*terraform.State) error {
		ds, err := testAccV2Client(t).Datasets.Get(context.Background(), environmentID, datasetID)
		if err != nil {
			return fmt.Errorf("could not read dataset %s: %w", datasetID, err)
		}
		if ds.Restricted == nil || *ds.Restricted != want {
			return fmt.Errorf("dataset %s: restricted is not %t", datasetID, want)
		}
		return nil
	}
}
