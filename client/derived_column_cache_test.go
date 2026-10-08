package client_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
)

// derivedColumnTestAPI is a fake of the derived columns API which counts
// the list and alias-lookup requests it serves.
type derivedColumnTestAPI struct {
	mu      sync.Mutex
	columns map[string][]client.DerivedColumn // keyed by dataset

	listRequests  atomic.Int64
	aliasRequests atomic.Int64
	writeRequests atomic.Int64
}

func (a *derivedColumnTestAPI) handler(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	defer a.mu.Unlock()

	dataset, id, _ := strings.Cut(strings.TrimPrefix(r.URL.Path, "/1/derived_columns/"), "/")

	switch {
	case r.Method == http.MethodGet && r.URL.Query().Has("alias"):
		a.aliasRequests.Add(1)
		for _, c := range a.columns[dataset] {
			if c.Alias == r.URL.Query().Get("alias") {
				_ = json.NewEncoder(w).Encode(c)
				return
			}
		}
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"derived column not found"}`))
	case r.Method == http.MethodGet:
		a.listRequests.Add(1)
		_ = json.NewEncoder(w).Encode(a.columns[dataset])
	case r.Method == http.MethodPost:
		a.writeRequests.Add(1)
		var c client.DerivedColumn
		_ = json.NewDecoder(r.Body).Decode(&c)
		c.ID = fmt.Sprintf("id-%d", len(a.columns[dataset])+1)
		a.columns[dataset] = append(a.columns[dataset], c)
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(c)
	case r.Method == http.MethodPut:
		a.writeRequests.Add(1)
		var c client.DerivedColumn
		_ = json.NewDecoder(r.Body).Decode(&c)
		c.ID = id
		for i := range a.columns[dataset] {
			if a.columns[dataset][i].ID == id {
				a.columns[dataset][i] = c
				_ = json.NewEncoder(w).Encode(c)
				return
			}
		}
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"derived column not found"}`))
	case r.Method == http.MethodDelete:
		a.writeRequests.Add(1)
		for i, c := range a.columns[dataset] {
			if c.ID == id {
				a.columns[dataset] = append(a.columns[dataset][:i:i], a.columns[dataset][i+1:]...)
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"derived column not found"}`))
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// testClientOption customizes the config of a test client.
type testClientOption func(*client.Config)

// withReadCaching enables read caching on a test client.
func withReadCaching(cfg *client.Config) { cfg.ReadCaching = true }

func newDerivedColumnTestClient(t *testing.T, api *derivedColumnTestAPI, opts ...testClientOption) *client.Client {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(api.handler))
	t.Cleanup(server.Close)

	cfg := &client.Config{
		APIKey: "test-key",
		APIUrl: server.URL,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	c, err := client.NewClientWithConfig(cfg)
	require.NoError(t, err)
	return c
}

func TestDerivedColumns_ReadsAreServedFromCache(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	api := &derivedColumnTestAPI{
		columns: map[string][]client.DerivedColumn{
			"test-dataset": {
				{ID: "id-1", Alias: "dc.one", Expression: "BOOL(1)"},
				{ID: "id-2", Alias: "dc.two", Expression: "BOOL(1)"},
			},
		},
	}
	c := newDerivedColumnTestClient(t, api, withReadCaching)

	// a burst of concurrent alias reads — as a refresh of many
	// honeycombio_derived_column resources produces — collapses into a
	// single list request
	var wg sync.WaitGroup
	for range 25 {
		for _, alias := range []string{"dc.one", "dc.two"} {
			wg.Add(1)
			go func() {
				defer wg.Done()
				dc, err := c.DerivedColumns.GetByAlias(ctx, "test-dataset", alias)
				assert.NoError(t, err)
				if assert.NotNil(t, dc) {
					assert.Equal(t, alias, dc.Alias)
				}
			}()
		}
	}
	wg.Wait()

	assert.Equal(t, int64(1), api.listRequests.Load(), "expected a single list request")
	assert.Zero(t, api.aliasRequests.Load(), "expected no per-alias requests")

	// List is served from the same cache
	columns, err := c.DerivedColumns.List(ctx, "test-dataset")
	require.NoError(t, err)
	assert.Len(t, columns, 2)
	assert.Equal(t, int64(1), api.listRequests.Load())
}

func TestDerivedColumns_MissingAliasFallsBackToDirectLookup(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	api := &derivedColumnTestAPI{
		columns: map[string][]client.DerivedColumn{"test-dataset": {}},
	}
	c := newDerivedColumnTestClient(t, api, withReadCaching)

	_, err := c.DerivedColumns.GetByAlias(ctx, "test-dataset", "dc.gone")

	var detailedErr client.DetailedError
	require.ErrorAs(t, err, &detailedErr)
	assert.True(t, detailedErr.IsNotFound(), "expected NotFound to propagate through the cache")
	assert.Equal(t, int64(1), api.aliasRequests.Load(), "expected the miss to be confirmed with a direct lookup")
}

func TestDerivedColumns_WritesInvalidateTheCache(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	api := &derivedColumnTestAPI{
		columns: map[string][]client.DerivedColumn{"test-dataset": {}},
	}
	c := newDerivedColumnTestClient(t, api, withReadCaching)

	// prime the cache with the empty dataset
	columns, err := c.DerivedColumns.List(ctx, "test-dataset")
	require.NoError(t, err)
	assert.Empty(t, columns)

	created, err := c.DerivedColumns.Create(ctx, "test-dataset", &client.DerivedColumn{
		Alias:      "dc.new",
		Expression: "BOOL(1)",
	})
	require.NoError(t, err)

	// the create invalidated the cached list, so the new column is
	// visible via a fresh list, without a per-alias lookup
	dc, err := c.DerivedColumns.GetByAlias(ctx, "test-dataset", "dc.new")
	require.NoError(t, err)
	assert.Equal(t, "dc.new", dc.Alias)
	assert.Equal(t, int64(2), api.listRequests.Load(), "expected the list to be refetched after the write")
	assert.Zero(t, api.aliasRequests.Load())

	// a delete also invalidates: the deleted column is gone from the
	// fresh list and the miss is confirmed with a direct lookup
	require.NoError(t, c.DerivedColumns.Delete(ctx, "test-dataset", created.ID))

	_, err = c.DerivedColumns.GetByAlias(ctx, "test-dataset", "dc.new")
	var detailedErr client.DetailedError
	require.ErrorAs(t, err, &detailedErr)
	assert.True(t, detailedErr.IsNotFound())
	assert.Equal(t, int64(3), api.listRequests.Load(), "expected the list to be refetched after the delete")
	assert.Equal(t, int64(1), api.aliasRequests.Load(), "expected the post-delete miss to be confirmed directly")
}

func TestDerivedColumns_CachingIsDisabledByDefault(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	api := &derivedColumnTestAPI{
		columns: map[string][]client.DerivedColumn{
			"test-dataset": {{ID: "id-1", Alias: "dc.one", Expression: "BOOL(1)"}},
		},
	}
	c := newDerivedColumnTestClient(t, api)

	// without read caching every alias read is its own direct lookup
	for range 2 {
		dc, err := c.DerivedColumns.GetByAlias(ctx, "test-dataset", "dc.one")
		require.NoError(t, err)
		assert.Equal(t, "dc.one", dc.Alias)
	}
	assert.Equal(t, int64(2), api.aliasRequests.Load(), "expected one direct lookup per read")
	assert.Zero(t, api.listRequests.Load(), "expected no list requests when caching is disabled")

	// and every list call hits the API
	for range 2 {
		_, err := c.DerivedColumns.List(ctx, "test-dataset")
		require.NoError(t, err)
	}
	assert.Equal(t, int64(2), api.listRequests.Load())
}

func TestDerivedColumns_EnvironmentWideWritesInvalidateAllDatasets(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	api := &derivedColumnTestAPI{
		columns: map[string][]client.DerivedColumn{
			"test-dataset":             {{ID: "id-1", Alias: "dc.one", Expression: "BOOL(1)"}},
			client.EnvironmentWideSlug: {},
		},
	}
	c := newDerivedColumnTestClient(t, api, withReadCaching)

	// prime the dataset-scoped cache
	_, err := c.DerivedColumns.List(ctx, "test-dataset")
	require.NoError(t, err)
	assert.Equal(t, int64(1), api.listRequests.Load())

	// an environment-wide write may be visible in every dataset's view,
	// so it drops the dataset-scoped cache too
	_, err = c.DerivedColumns.Create(ctx, client.EnvironmentWideSlug, &client.DerivedColumn{
		Alias:      "dc.env_wide",
		Expression: "BOOL(1)",
	})
	require.NoError(t, err)

	_, err = c.DerivedColumns.List(ctx, "test-dataset")
	require.NoError(t, err)
	assert.Equal(t, int64(2), api.listRequests.Load(), "expected the dataset list to be refetched after an env-wide write")
}

func TestDerivedColumns_ListErrorFallsBackToDirectLookup(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// a server whose list endpoint always fails (with a non-retryable
	// status) but can serve alias lookups
	var aliasRequests atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Has("alias") {
			aliasRequests.Add(1)
			_ = json.NewEncoder(w).Encode(client.DerivedColumn{
				ID: "id-1", Alias: r.URL.Query().Get("alias"), Expression: "BOOL(1)",
			})
			return
		}
		w.WriteHeader(http.StatusBadRequest)
	}))
	t.Cleanup(server.Close)

	c, err := client.NewClientWithConfig(&client.Config{
		APIKey:      "test-key",
		APIUrl:      server.URL,
		ReadCaching: true,
	})
	require.NoError(t, err)

	dc, err := c.DerivedColumns.GetByAlias(ctx, "test-dataset", "dc.one")
	require.NoError(t, err)
	assert.Equal(t, "dc.one", dc.Alias)
	assert.Equal(t, int64(1), aliasRequests.Load())

	// List propagates the error
	_, err = c.DerivedColumns.List(ctx, "test-dataset")
	require.Error(t, err)
	var detailedErr client.DetailedError
	require.ErrorAs(t, err, &detailedErr)
}
