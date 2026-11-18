package v2

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	hnyclient "github.com/honeycombio/terraform-provider-honeycombio/client"
)

const (
	testGroupID   = "hcxgr_01jbr2v3k8m4n5p6q7r8s9t0vw"
	testGroupPath = "/2/teams/" + testTeamSlug + "/groups/" + testGroupID

	testGroupJSON = `{"data": {"id": "` + testGroupID + `", "type": "groups",
		"links": {"self": "` + testGroupPath + `"},
		"attributes": {"name": "payments-oncall", "description": "Reads payments",
			"member_count": 3,
			"timestamps": {"created": "2026-09-22T17:32:11Z", "updated": "2026-10-31T15:08:11Z"}}}}`
)

func TestClient_Groups(t *testing.T) {
	ctx := context.Background()

	t.Run("create sends the name and description and reads the Group back", func(t *testing.T) {
		c, f := newFakeAPIClient(t, map[string]http.HandlerFunc{
			"POST /2/teams/" + testTeamSlug + "/groups": func(w http.ResponseWriter, _ *http.Request) {
				writeJSONAPI(w, http.StatusCreated, testGroupJSON)
			},
		})

		g, err := c.Groups.Create(ctx, &Group{Name: "payments-oncall", Description: new("")})
		require.NoError(t, err)

		reqs := f.recorded()
		require.Len(t, reqs, 1)
		// an empty description is sent, so an update can clear it;
		// read-only attributes are never sent
		assert.JSONEq(t,
			`{"data":{"type":"groups","attributes":{"name":"payments-oncall","description":""}}}`,
			reqs[0].Body,
		)

		assert.Equal(t, testGroupID, g.ID)
		assert.Equal(t, "payments-oncall", g.Name)
		assert.Equal(t, "Reads payments", *g.Description)
		assert.Equal(t, 3, g.MemberCount)
		require.NotNil(t, g.Timestamps)
		assert.Equal(t, time.Date(2026, 9, 22, 17, 32, 11, 0, time.UTC), g.Timestamps.CreatedAt)
	})

	t.Run("update patches the Group at its own path", func(t *testing.T) {
		c, f := newFakeAPIClient(t, map[string]http.HandlerFunc{
			"PATCH " + testGroupPath: func(w http.ResponseWriter, _ *http.Request) {
				writeJSONAPI(w, http.StatusOK, testGroupJSON)
			},
		})

		_, err := c.Groups.Update(ctx, &Group{ID: testGroupID, Name: "renamed", Description: new("new")})
		require.NoError(t, err)

		reqs := f.recorded()
		require.Len(t, reqs, 1)
		assert.JSONEq(t,
			`{"data":{"type":"groups","id":"`+testGroupID+`","attributes":{"name":"renamed","description":"new"}}}`,
			reqs[0].Body,
		)
	})

	t.Run("get and delete report a missing Group as not found", func(t *testing.T) {
		notFound := func(w http.ResponseWriter, _ *http.Request) {
			writeJSONAPI(w, http.StatusNotFound,
				`{"errors":[{"status":"404","code":"resource/not-found","title":"The requested resource cannot be found."}]}`)
		}
		c, _ := newFakeAPIClient(t, map[string]http.HandlerFunc{
			"GET " + testGroupPath:    notFound,
			"DELETE " + testGroupPath: notFound,
		})

		var de hnyclient.DetailedError
		_, err := c.Groups.Get(ctx, testGroupID)
		require.ErrorAs(t, err, &de)
		assert.True(t, de.IsNotFound())

		err = c.Groups.Delete(ctx, testGroupID)
		require.ErrorAs(t, err, &de)
		assert.True(t, de.IsNotFound())
	})

	t.Run("replace members sends identifiers only, and no users as an empty list", func(t *testing.T) {
		c, f := newFakeAPIClient(t, map[string]http.HandlerFunc{
			"PATCH " + testGroupPath + "/relationships/members": noContent,
		})

		require.NoError(t, c.Groups.ReplaceMembers(ctx, testGroupID, []string{"hcxus_a", "hcxus_b"}))
		// the API rejects a missing or null list, so an empty one must be sent as []
		require.NoError(t, c.Groups.ReplaceMembers(ctx, testGroupID, []string{}))

		reqs := f.recorded()
		require.Len(t, reqs, 2)
		assert.JSONEq(t, `{"data":[{"type":"users","id":"hcxus_a"},{"type":"users","id":"hcxus_b"}]}`, reqs[0].Body)
		assert.JSONEq(t, `{"data":[]}`, reqs[1].Body)
	})

	t.Run("list members follows the next link to the last page", func(t *testing.T) {
		c, _ := newFakeAPIClient(t, map[string]http.HandlerFunc{
			"GET " + testGroupPath + "/relationships/members": func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Query().Get("page[after]") == "" {
					writeJSONAPI(w, http.StatusOK, `{
						"data": [{"type": "users", "id": "hcxus_a"}, {"type": "users", "id": "hcxus_b"}],
						"links": {"next": "`+testGroupPath+`/relationships/members?page%5Bafter%5D=hcxus_b&page%5Bsize%5D=2"}}`)
					return
				}
				writeJSONAPI(w, http.StatusOK, `{"data": [{"type": "users", "id": "hcxus_c"}], "links": {}}`)
			},
		})

		pager, err := c.Groups.ListMembers(ctx, testGroupID, PageSize(2))
		require.NoError(t, err)
		var ids []string
		for pager.HasNext() {
			items, err := pager.Next(ctx)
			require.NoError(t, err)
			for _, m := range items {
				ids = append(ids, m.ID)
			}
		}
		assert.Equal(t, []string{"hcxus_a", "hcxus_b", "hcxus_c"}, ids)
	})

	datasetIDs := make([]string, 250)
	for i := range datasetIDs {
		datasetIDs[i] = fmt.Sprintf("hcxds_%03d", i)
	}
	batchIDs := func(t *testing.T, body string) []string {
		t.Helper()
		var req struct {
			Data []struct {
				Type string `json:"type"`
				ID   string `json:"id"`
			} `json:"data"`
		}
		require.NoError(t, json.Unmarshal([]byte(body), &req))
		ids := make([]string, len(req.Data))
		for i, d := range req.Data {
			assert.Equal(t, "datasets", d.Type)
			ids[i] = d.ID
		}
		return ids
	}

	t.Run("grant and revoke send batches of 100 in order", func(t *testing.T) {
		datasetsPath := testGroupPath + "/relationships/datasets"
		c, f := newFakeAPIClient(t, map[string]http.HandlerFunc{
			"POST " + datasetsPath:   noContent,
			"DELETE " + datasetsPath: noContent,
		})

		require.NoError(t, c.Groups.GrantDatasets(ctx, testGroupID, datasetIDs))
		require.NoError(t, c.Groups.RevokeDatasets(ctx, testGroupID, datasetIDs[:1]))

		reqs := f.recorded()
		require.Len(t, reqs, 4)
		for i, want := range [][]string{datasetIDs[0:100], datasetIDs[100:200], datasetIDs[200:250]} {
			assert.Equal(t, http.MethodPost, reqs[i].Method)
			assert.Equal(t, want, batchIDs(t, reqs[i].Body), "batch %d", i)
		}
		assert.Equal(t, http.MethodDelete, reqs[3].Method)
		assert.JSONEq(t, `{"data":[{"type":"datasets","id":"hcxds_000"}]}`, reqs[3].Body)
	})

	t.Run("grant stops at the first failed batch", func(t *testing.T) {
		var calls int
		c, f := newFakeAPIClient(t, map[string]http.HandlerFunc{
			"POST " + testGroupPath + "/relationships/datasets": func(w http.ResponseWriter, _ *http.Request) {
				calls++
				if calls == 2 {
					writeJSONAPI(w, http.StatusUnprocessableEntity, `{"errors":[{"status":"422",
						"code":"validation-failed/invalid-field","title":"The request failed validation.",
						"detail":"only a restricted dataset can be granted to a group",
						"source":{"pointer":"/data/7"}}]}`)
					return
				}
				w.WriteHeader(http.StatusNoContent)
			},
		})

		err := c.Groups.GrantDatasets(ctx, testGroupID, datasetIDs)
		var de hnyclient.DetailedError
		require.ErrorAs(t, err, &de)
		assert.Equal(t, http.StatusUnprocessableEntity, de.Status)
		assert.Contains(t, de.Error(), "only a restricted dataset can be granted")
		assert.Len(t, f.recorded(), 2, "no batch is sent after a failure")
	})
}
