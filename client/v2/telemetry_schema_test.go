package v2

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/hashicorp/jsonapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	hnyclient "github.com/honeycombio/terraform-provider-honeycombio/client"
)

const testTeamSlug = "test-team"

// fakeTelemetrySchemaServer is a minimal stand-in for the telemetry schema routes.
// It stores the decoded schema and replays it, the way the real API
// is documented to, so the tests can check the client alters nothing in between.
//
// It builds its responses with encoding/json rather than the jsonapi package
// the client uses, so a bug in the client's encoding can't cancel itself out.
type fakeTelemetrySchemaServer struct {
	*httptest.Server

	mu       sync.Mutex
	schema   *string
	requests []*http.Request
	bodies   []string
}

func newFakeTelemetrySchemaServer(t *testing.T) *fakeTelemetrySchemaServer {
	t.Helper()

	f := &fakeTelemetrySchemaServer{}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /2/auth", func(w http.ResponseWriter, r *http.Request) {
		f.writeJSON(w, http.StatusOK, `{
			"data": {"id": "key1", "type": "api-keys",
				"attributes": {"name": "k", "key_type": "management", "disabled": false, "scopes": []},
				"relationships": {"team": {"data": {"type": "teams", "id": "team1"}}}},
			"included": [{"id": "team1", "type": "teams", "attributes": {"name": "Test", "slug": "`+testTeamSlug+`"}}]
		}`)
	})
	mux.HandleFunc("GET /2/teams/"+testTeamSlug+"/telemetry-schema", func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		if f.schema == nil {
			f.writeJSON(w, http.StatusNotFound, `{"errors":[{"status":"404","code":"resource/not-found","title":"The requested resource cannot be found."}]}`)
			return
		}
		f.writeSchema(w, *f.schema)
	})
	mux.HandleFunc("PUT /2/teams/"+testTeamSlug+"/telemetry-schema", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Data struct {
				Type       string `json:"type"`
				Attributes struct {
					SchemaYAML string `json:"schema_yaml"`
				} `json:"attributes"`
			} `json:"data"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Data.Type != "telemetry-schema" {
			f.writeJSON(w, http.StatusBadRequest, `{"errors":[{"status":"400","code":"request/invalid","title":"bad request"}]}`)
			return
		}
		if req.Data.Attributes.SchemaYAML == "" {
			f.writeJSON(w, http.StatusUnprocessableEntity, `{"errors":[{"status":"422","code":"field/invalid","title":"The request failed validation.","detail":"cannot be blank","source":{"pointer":"/data/attributes/schema_yaml"}}]}`)
			return
		}
		f.mu.Lock()
		defer f.mu.Unlock()
		f.schema = &req.Data.Attributes.SchemaYAML
		f.writeSchema(w, *f.schema)
	})
	mux.HandleFunc("DELETE /2/teams/"+testTeamSlug+"/telemetry-schema", func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		f.schema = nil
		w.WriteHeader(http.StatusNoContent)
	})

	f.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		r.Body = io.NopCloser(bytes.NewReader(body))
		f.mu.Lock()
		f.requests = append(f.requests, r)
		f.bodies = append(f.bodies, string(body))
		f.mu.Unlock()
		mux.ServeHTTP(w, r)
	}))
	t.Cleanup(f.Close)
	return f
}

func (f *fakeTelemetrySchemaServer) writeJSON(w http.ResponseWriter, status int, body string) {
	w.Header().Set("Content-Type", jsonapi.MediaType)
	w.WriteHeader(status)
	_, _ = io.WriteString(w, body)
}

func (f *fakeTelemetrySchemaServer) writeSchema(w http.ResponseWriter, schema string) {
	body, _ := json.Marshal(map[string]any{"data": map[string]any{
		"id":         testTeamSlug,
		"type":       "telemetry-schema",
		"links":      map[string]string{"self": "/2/teams/" + testTeamSlug + "/telemetry-schema"},
		"attributes": map[string]string{"schema_yaml": schema},
	}})
	w.Header().Set("ETag", `"abc"`)
	f.writeJSON(w, http.StatusOK, string(body))
}

func TestClient_TelemetrySchemas(t *testing.T) {
	ctx := context.Background()

	newClient := func(t *testing.T) (*Client, *fakeTelemetrySchemaServer) {
		t.Helper()
		f := newFakeTelemetrySchemaServer(t)
		c, err := NewClientWithConfig(&Config{
			APIKeyID:     "id",
			APIKeySecret: "secret",
			BaseURL:      f.URL,
			UserAgent:    testUserAgent,
		})
		require.NoError(t, err)
		return c, f
	}

	t.Run("round trip is byte-identical", func(t *testing.T) {
		// Each document is one a normalising client would change.
		docs := map[string]string{
			"plain":                   "groups:\n  - id: registry.myapp\n    type: attribute_group\n",
			"no trailing newline":     "groups: []",
			"several newlines":        "groups: []\n\n\n",
			"CRLF line endings":       "groups:\r\n  - id: a\r\n",
			"trailing whitespace":     "groups:   \n  - id: a  \n\t\n",
			"comments and odd indent": "# keep me\ngroups:\n     - id:    a   # trailing\n",
			"JSON-escaped chars":      "brief: \"quotes \\\" backslash \\\\ <tag> & \u2028 \u2029 \x7f\"\n",
			"non-ASCII":               "brief: héllo wörld 日本語 🙂\n",
			"leading BOM":             "\ufeffgroups: []\n",
			"leading blank lines":     "\n\n  groups: []\n",
		}
		for name, doc := range docs {
			t.Run(name, func(t *testing.T) {
				c, f := newClient(t)

				put, err := c.TelemetrySchemas.Put(ctx, &TelemetrySchema{SchemaYAML: doc})
				require.NoError(t, err)
				assert.Equal(t, doc, put.SchemaYAML)
				assert.Equal(t, testTeamSlug, put.ID)

				got, err := c.TelemetrySchemas.Get(ctx)
				require.NoError(t, err)
				assert.Equal(t, doc, got.SchemaYAML)
				assert.Equal(t, testTeamSlug, got.ID)

				// the document must also be what was on the wire
				f.mu.Lock()
				defer f.mu.Unlock()
				require.NotNil(t, f.schema)
				assert.Equal(t, doc, *f.schema)
			})
		}
	})

	t.Run("sends the expected requests", func(t *testing.T) {
		c, f := newClient(t)

		_, err := c.TelemetrySchemas.Put(ctx, &TelemetrySchema{SchemaYAML: "groups: []\n"})
		require.NoError(t, err)

		f.mu.Lock()
		defer f.mu.Unlock()
		// the team slug lookup comes first, so the PUT is the last request
		put := f.requests[len(f.requests)-1]
		assert.Equal(t, http.MethodPut, put.Method)
		assert.Equal(t, "/2/teams/"+testTeamSlug+"/telemetry-schema", put.URL.Path)
		assert.Equal(t, "Bearer id:secret", put.Header.Get("Authorization"))
		assert.Equal(t, jsonapi.MediaType, put.Header.Get("Content-Type"))
		assert.JSONEq(t,
			`{"data":{"type":"telemetry-schema","attributes":{"schema_yaml":"groups: []\n"}}}`,
			f.bodies[len(f.bodies)-1],
		)
	})

	t.Run("get returns not found when no schema is set", func(t *testing.T) {
		c, _ := newClient(t)

		_, err := c.TelemetrySchemas.Get(ctx)
		var de hnyclient.DetailedError
		require.ErrorAs(t, err, &de)
		assert.True(t, de.IsNotFound())
	})

	t.Run("put surfaces validation errors", func(t *testing.T) {
		c, _ := newClient(t)

		_, err := c.TelemetrySchemas.Put(ctx, &TelemetrySchema{})
		var de hnyclient.DetailedError
		require.ErrorAs(t, err, &de)
		assert.Equal(t, http.StatusUnprocessableEntity, de.Status)
		assert.Contains(t, de.Error(), "cannot be blank")
	})

	t.Run("put replaces the existing schema", func(t *testing.T) {
		c, _ := newClient(t)

		_, err := c.TelemetrySchemas.Put(ctx, &TelemetrySchema{SchemaYAML: "a: 1\n"})
		require.NoError(t, err)
		_, err = c.TelemetrySchemas.Put(ctx, &TelemetrySchema{SchemaYAML: "b: 2\n"})
		require.NoError(t, err)

		got, err := c.TelemetrySchemas.Get(ctx)
		require.NoError(t, err)
		assert.Equal(t, "b: 2\n", got.SchemaYAML)
	})

	t.Run("delete removes the schema and succeeds when none is set", func(t *testing.T) {
		c, _ := newClient(t)

		require.NoError(t, c.TelemetrySchemas.Delete(ctx), "nothing to delete")

		_, err := c.TelemetrySchemas.Put(ctx, &TelemetrySchema{SchemaYAML: "a: 1\n"})
		require.NoError(t, err)
		require.NoError(t, c.TelemetrySchemas.Delete(ctx))

		_, err = c.TelemetrySchemas.Get(ctx)
		var de hnyclient.DetailedError
		require.ErrorAs(t, err, &de)
		assert.True(t, de.IsNotFound())
	})
}
