package v2

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_Datasets(t *testing.T) {
	ctx := context.Background()
	const (
		envID       = "hcaen_01j1jrsewaha3m0z6fwffpcrxg"
		datasetID   = "hcxds_01jbr4k2m7n8p9q0r1s2t3v4wx"
		datasetPath = "/2/teams/" + testTeamSlug + "/environments/" + envID + "/datasets/" + datasetID
		datasetJSON = `{"data": {"id": "` + datasetID + `", "type": "datasets",
			"attributes": {"name": "Checkout", "slug": "checkout", "description": "",
				"expand_json_depth": 0, "settings": {"delete_protected": true},
				"dataset_type": "events", "restricted": true, "regular_columns_count": 0,
				"last_written_at": null, "created_at": "2026-11-18T15:08:25Z"},
			"relationships": {"environment": {"data": {"type": "environments", "id": "` + envID + `"}}}}}`
	)

	c, f := newFakeAPIClient(t, map[string]http.HandlerFunc{
		"GET " + datasetPath: func(w http.ResponseWriter, _ *http.Request) {
			writeJSONAPI(w, http.StatusOK, datasetJSON)
		},
		"PATCH " + datasetPath: func(w http.ResponseWriter, _ *http.Request) {
			writeJSONAPI(w, http.StatusOK, datasetJSON)
		},
	})

	t.Run("get reads every field, and a null last_written_at as nil", func(t *testing.T) {
		ds, err := c.Datasets.Get(ctx, envID, datasetID)
		require.NoError(t, err)
		assert.Equal(t, datasetID, ds.ID)
		assert.Equal(t, "checkout", ds.Slug)
		assert.Equal(t, 0, *ds.ExpandJSONDepth)
		assert.True(t, *ds.Settings.DeleteProtected)
		assert.True(t, *ds.Restricted)
		assert.Equal(t, "events", ds.DatasetType)
		assert.Nil(t, ds.LastWrittenAt)
		assert.Equal(t, time.Date(2026, 11, 18, 15, 8, 25, 0, time.UTC), *ds.CreatedAt)
	})

	t.Run("update sends only the fields that are set, including false", func(t *testing.T) {
		_, err := c.Datasets.Update(ctx, envID, &Dataset{ID: datasetID, Restricted: new(false)})
		require.NoError(t, err)

		reqs := f.recorded()
		last := reqs[len(reqs)-1]
		assert.Equal(t, http.MethodPatch, last.Method)
		assert.JSONEq(t,
			`{"data":{"type":"datasets","id":"`+datasetID+`","attributes":{"restricted":false}}}`,
			last.Body,
		)
	})
}
