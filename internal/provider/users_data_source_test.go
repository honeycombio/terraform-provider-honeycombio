package provider

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/test"
)

func TestAcc_UsersDataSource(t *testing.T) {
	users := testAccTeamUsers(t, 3)
	name := test.RandomStringWithPrefix("test.", 20)

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheckV2API(t),
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactory,
		CheckDestroy:             testAccEnsureGroupDestroyed(t),
		Steps: []resource.TestStep{
			{ // without group_id, every user on the Team
				Config: `data "honeycombio_users" "test" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("data.honeycombio_users.test", "users.#", func(v string) error {
						if n, _ := strconv.Atoi(v); n < len(users) {
							return fmt.Errorf("expected at least %d users, got %s", len(users), v)
						}
						return nil
					}),
					resource.TestCheckTypeSetElemNestedAttrs("data.honeycombio_users.test", "users.*", map[string]string{
						"id":    users[0].ID,
						"email": users[0].Email,
						"role":  users[0].Role,
					}),
				),
			},
			{ // with group_id, only the Group's users
				Config: testAccConfigGroupMembers(name, users[1].ID, users[2].ID) + `

data "honeycombio_users" "test" {
  group_id = honeycombio_group_members.test.group_id
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.honeycombio_users.test", "users.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("data.honeycombio_users.test", "users.*", map[string]string{
						"id":    users[1].ID,
						"email": users[1].Email,
					}),
					resource.TestCheckTypeSetElemNestedAttrs("data.honeycombio_users.test", "users.*", map[string]string{
						"id":    users[2].ID,
						"email": users[2].Email,
					}),
				),
			},
		},
	})
}

// TestUsersDataSource_fakeAPI covers what a real Team can't show cheaply:
// more than one page of users, and the errors for a 404.
func TestUsersDataSource_fakeAPI(t *testing.T) {
	t.Run("follows every page, and keeps the group filter", func(t *testing.T) {
		var queries []url.Values
		testAccFakeUsersAPI(t, func(w http.ResponseWriter, r *http.Request) {
			queries = append(queries, r.URL.Query())
			if r.URL.Query().Get("page[after]") == "" {
				testWriteJSONAPI(w, http.StatusOK, `{"data": [
					{"type": "users", "id": "hcxus_a", "attributes": {"email": "a@example.com", "role": "owner"}},
					{"type": "users", "id": "hcxus_b", "attributes": {"email": "b@example.com", "role": "member"}}],
					"links": {"next": "/2/teams/fake/users?page%5Bafter%5D=hcxus_b&page%5Bsize%5D=100&filter%5Bgroup%5D=hcxgr_1"}}`)
				return
			}
			testWriteJSONAPI(w, http.StatusOK, `{"data": [
				{"type": "users", "id": "hcxus_c", "attributes": {"email": "c@example.com", "role": "read-only"}}],
				"links": {}}`)
		})

		resource.UnitTest(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactory,
			Steps: []resource.TestStep{{
				Config: `data "honeycombio_users" "test" { group_id = "hcxgr_1" }`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.honeycombio_users.test", "users.#", "3"),
					resource.TestCheckResourceAttr("data.honeycombio_users.test", "users.0.id", "hcxus_a"),
					resource.TestCheckResourceAttr("data.honeycombio_users.test", "users.2.id", "hcxus_c"),
					resource.TestCheckResourceAttr("data.honeycombio_users.test", "users.2.role", "read-only"),
					func(*terraform.State) error {
						for _, q := range queries {
							if q.Get("filter[group]") != "hcxgr_1" {
								return fmt.Errorf("a request did not filter by the Group: %v", q)
							}
						}
						return nil
					},
				),
			}},
		})
	})

	t.Run("an unknown Group gets an error on group_id", func(t *testing.T) {
		testAccFakeUsersAPI(t, testNotFound)

		resource.UnitTest(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactory,
			Steps: []resource.TestStep{{
				Config:      `data "honeycombio_users" "test" { group_id = "hcxgr_missing" }`,
				ExpectError: regexp.MustCompile(`No Group has the ID "hcxgr_missing"`),
			}},
		})
	})

	t.Run("a 404 without group_id says the Team can't list users", func(t *testing.T) {
		testAccFakeUsersAPI(t, testNotFound)

		resource.UnitTest(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactory,
			Steps: []resource.TestStep{
				{
					Config:      `data "honeycombio_users" "test" {}`,
					ExpectError: regexp.MustCompile(`This Team can't list users`),
				},
				{
					Config:      `data "honeycombio_user" "test" { email = "a@example.com" }`,
					ExpectError: regexp.MustCompile(`This Team can't list users`),
				},
			},
		})
	})
}

// testAccFakeUsersAPI points the provider at a local fake of the v2 API for
// the rest of the test. users serves GET /2/teams/fake/users.
func testAccFakeUsersAPI(t *testing.T, users http.HandlerFunc) {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /2/auth", func(w http.ResponseWriter, _ *http.Request) {
		testWriteJSONAPI(w, http.StatusOK, `{
			"data": {"id": "k", "type": "api-keys",
				"attributes": {"name": "k", "key_type": "management", "scopes": ["groups:write", "users:read"]},
				"relationships": {"team": {"data": {"type": "teams", "id": "t"}}}},
			"included": [{"id": "t", "type": "teams", "attributes": {"name": "Fake", "slug": "fake"}}]}`)
	})
	mux.HandleFunc("GET /2/teams/fake/users", users)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	t.Setenv("HONEYCOMB_API_ENDPOINT", srv.URL)
	t.Setenv("HONEYCOMB_API_KEY", "")
	t.Setenv("HONEYCOMB_KEY_ID", "fake")
	t.Setenv("HONEYCOMB_KEY_SECRET", "fake")
}

func testNotFound(w http.ResponseWriter, _ *http.Request) {
	testWriteJSONAPI(w, http.StatusNotFound, `{"errors": [{"status": "404",
		"code": "resource/not-found", "title": "The requested resource cannot be found."}]}`)
}

func testWriteJSONAPI(w http.ResponseWriter, status int, body string) {
	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.WriteHeader(status)
	_, _ = io.WriteString(w, body)
}
