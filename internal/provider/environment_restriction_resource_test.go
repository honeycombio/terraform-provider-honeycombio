package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"

	v2client "github.com/honeycombio/terraform-provider-honeycombio/client/v2"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/test"
)

func TestAcc_EnvironmentRestrictionResource(t *testing.T) {
	envID := testAccEnvironment(context.Background(), t, testAccV2Client(t)).ID
	config := testAccConfigEnvironmentRestriction(envID)

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheckV2API(t),
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactory,
		// the Environment stays, without its restriction
		CheckDestroy: testAccEnsureEnvironmentRestricted(t, envID, false),
		Steps: []resource.TestStep{
			{ // the documented order: grant the Environment first, then restrict it
				Config: testAccConfigGroup(test.RandomStringWithPrefix("test.", 20), "") + fmt.Sprintf(`
resource "honeycombio_group_grant" "test" {
  group_id       = honeycombio_group.test.id
  environment_id = "%[1]s"
}

resource "honeycombio_environment_restriction" "test" {
  environment_id = "%[1]s"

  depends_on = [honeycombio_group_grant.test]
}

data "honeycombio_environment" "test" {
  id = honeycombio_environment_restriction.test.environment_id
}

data "honeycombio_environments" "restricted" {
  detail_filter {
    name  = "restricted"
    value = "true"
  }

  depends_on = [honeycombio_environment_restriction.test]
}`, envID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("honeycombio_environment_restriction.test", "id", envID),
					resource.TestCheckResourceAttr("honeycombio_environment_restriction.test", "environment_id", envID),
					testAccEnsureEnvironmentRestricted(t, envID, true),
					resource.TestCheckResourceAttr("data.honeycombio_environment.test", "restricted", "true"),
					resource.TestCheckTypeSetElemAttr("data.honeycombio_environments.restricted", "ids.*", envID),
					// restricting keeps the grant, so the Group can still read the Environment
					testAccEnsureGroupGrant(t, environmentTarget, envID, true),
				),
			},
			{
				ResourceName:      "honeycombio_environment_restriction.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{ // a restriction removed outside Terraform is created again
				PreConfig: func() {
					testAccSetEnvironmentRestricted(t, envID, false)
				},
				Config: config,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("honeycombio_environment_restriction.test", plancheck.ResourceActionCreate),
					},
				},
				Check: testAccEnsureEnvironmentRestricted(t, envID, true),
			},
		},
	})
}

// An Environment restricted outside Terraform, such as in the UI, must be imported,
// so that deleting the resource can't remove a restriction it didn't create.
func TestAcc_EnvironmentRestrictionResource_alreadyRestricted(t *testing.T) {
	envID := testAccEnvironment(context.Background(), t, testAccV2Client(t)).ID
	testAccSetEnvironmentRestricted(t, envID, true)

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheckV2API(t),
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactory,
		CheckDestroy:             testAccEnsureEnvironmentRestricted(t, envID, true),
		Steps: []resource.TestStep{
			{
				Config:      testAccConfigEnvironmentRestriction(envID),
				ExpectError: regexp.MustCompile(`Honeycomb Environment Is Already Restricted`),
			},
		},
	})
}

func TestAcc_EnvironmentRestrictionResource_environmentDeleted(t *testing.T) {
	ctx := context.Background()
	c := testAccV2Client(t)
	envID := testAccEnvironment(ctx, t, c).ID

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheckV2API(t),
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactory,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigEnvironmentRestriction(envID),
			},
			{ // the restriction of a deleted Environment leaves state, so the plan creates it again
				PreConfig: func() {
					_, err := c.Environments.Update(ctx, &v2client.Environment{
						ID:       envID,
						Settings: &v2client.EnvironmentSettings{DeleteProtected: helper.ToPtr(false)},
					})
					require.NoError(t, err)
					require.NoError(t, c.Environments.Delete(ctx, envID))
				},
				RefreshState:       true,
				ExpectNonEmptyPlan: true,
			},
			{
				Config:      testAccConfigEnvironmentRestriction(envID),
				ExpectError: regexp.MustCompile(`The API found no Environment "` + envID + `"`),
			},
		},
	})
}

func TestAcc_EnvironmentRestrictionResource_errors(t *testing.T) {
	envID := testAccEnvironment(context.Background(), t, testAccV2Client(t)).ID

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheckV2API(t),
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactory,
		Steps: []resource.TestStep{
			{ // an Environment that is not restricted has no restriction to import
				Config:        testAccConfigEnvironmentRestriction(envID),
				ResourceName:  "honeycombio_environment_restriction.test",
				ImportState:   true,
				ImportStateId: envID,
				ExpectError:   regexp.MustCompile(`Cannot import non-existent remote object`),
			},
		},
	})
}

// TestEnvironmentRestrictionResource_fakeAPI covers the API's 403 for a key
// that can't revoke the Dataset grants in the Environment. An acceptance test
// would need a second Management Key.
func TestEnvironmentRestrictionResource_fakeAPI(t *testing.T) {
	const (
		envID   = "hcxen_01j1jrsewaha3m0z6fwffpcrxg"
		envPath = "/2/teams/fake/environments/" + envID
	)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /2/auth", func(w http.ResponseWriter, _ *http.Request) {
		testWriteJSONAPI(w, http.StatusOK, `{
			"data": {"id": "k", "type": "api-keys",
				"attributes": {"name": "k", "key_type": "management", "scopes": ["environments:write"]},
				"relationships": {"team": {"data": {"type": "teams", "id": "t"}}}},
			"included": [{"id": "t", "type": "teams", "attributes": {"name": "Fake", "slug": "fake"}}]}`)
	})
	mux.HandleFunc("GET "+envPath, func(w http.ResponseWriter, _ *http.Request) {
		testWriteJSONAPI(w, http.StatusOK, `{"data": {"id": "`+envID+`", "type": "environments",
			"attributes": {"name": "prod", "slug": "prod", "restricted": false}}}`)
	})
	mux.HandleFunc("PATCH "+envPath, func(w http.ResponseWriter, _ *http.Request) {
		// the API's response to a key that can't revoke the Dataset grants in the Environment
		testWriteJSONAPI(w, http.StatusForbidden, `{"errors": [{"id": "d5c966612468262cb8e4048e41dce194",
			"title": "this API key is not allowed to access this resource", "code": "permission-denied/not-allowed"}]}`)
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
				Config:      testAccConfigEnvironmentRestriction(envID),
				ExpectError: regexp.MustCompile(`(?s)not allowed to access this resource.*cannot restrict this Environment`),
			},
		},
	})
}

func testAccConfigEnvironmentRestriction(environmentID string) string {
	return fmt.Sprintf(`
resource "honeycombio_environment_restriction" "test" {
  environment_id = "%s"
}`, environmentID)
}

func testAccSetEnvironmentRestricted(t *testing.T, environmentID string, restricted bool) {
	t.Helper()

	_, err := testAccV2Client(t).Environments.Update(context.Background(), &v2client.Environment{
		ID:         environmentID,
		Restricted: &restricted,
	})
	require.NoError(t, err)
}

func testAccEnsureEnvironmentRestricted(t *testing.T, environmentID string, want bool) resource.TestCheckFunc {
	return func(*terraform.State) error {
		env, err := testAccV2Client(t).Environments.Get(context.Background(), environmentID)
		if err != nil {
			return fmt.Errorf("could not read Environment %s: %w", environmentID, err)
		}
		if env.Restricted == nil || *env.Restricted != want {
			return fmt.Errorf("Environment %s: restricted is not %t", environmentID, want)
		}
		return nil
	}
}
