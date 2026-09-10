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

// columnTestAPI is a fake of the columns API which counts the list and
// key-name-lookup requests it serves.
type columnTestAPI struct {
	mu      sync.Mutex
	columns map[string][]client.Column // keyed by dataset

	listRequests    atomic.Int64
	keyNameRequests atomic.Int64
	writeRequests   atomic.Int64

	nextID int
}

func (a *columnTestAPI) handler(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// path is "{dataset}" for the collection, "{dataset}/{id}" for an item
	rest := strings.TrimPrefix(r.URL.Path, "/1/columns/")
	dataset, id, _ := strings.Cut(rest, "/")

	switch {
	case r.Method == http.MethodGet && r.URL.Query().Has("key_name"):
		a.keyNameRequests.Add(1)
		for _, c := range a.columns[dataset] {
			if c.KeyName == r.URL.Query().Get("key_name") {
				_ = json.NewEncoder(w).Encode(c)
				return
			}
		}
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"column not found"}`))

	case r.Method == http.MethodGet && id == "":
		a.listRequests.Add(1)
		_ = json.NewEncoder(w).Encode(a.columns[dataset])

	case r.Method == http.MethodPost:
		a.writeRequests.Add(1)
		var c client.Column
		_ = json.NewDecoder(r.Body).Decode(&c)
		a.nextID++
		c.ID = fmt.Sprintf("id-%d", a.nextID)
		a.columns[dataset] = append(a.columns[dataset], c)
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(c)

	case r.Method == http.MethodPut:
		a.writeRequests.Add(1)
		var c client.Column
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
		_, _ = w.Write([]byte(`{"error":"column not found"}`))

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
		_, _ = w.Write([]byte(`{"error":"column not found"}`))

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func newColumnTestClient(t *testing.T, api *columnTestAPI, opts ...testClientOption) *client.Client {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(api.handler))
	t.Cleanup(server.Close)

	cfg := &client.Config{APIKey: "test-key", APIUrl: server.URL}
	for _, o := range opts {
		o(cfg)
	}
	c, err := client.NewClientWithConfig(cfg)
	require.NoError(t, err)
	return c
}

func TestColumns_ReadsAreServedFromCache(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	api := &columnTestAPI{
		columns: map[string][]client.Column{
			"test-dataset": {
				{ID: "id-1", KeyName: "column_1"},
				{ID: "id-2", KeyName: "column_2"},
			},
		},
	}
	c := newColumnTestClient(t, api, withReadCaching)

	// a burst of concurrent key-name reads -- as a refresh of many
	// honeycombio_column resources produces -- collapses into a single
	// list request
	var wg sync.WaitGroup
	for range 25 {
		for _, keyName := range []string{"column_1", "column_2"} {
			wg.Add(1)
			go func() {
				defer wg.Done()
				col, err := c.Columns.GetByKeyName(ctx, "test-dataset", keyName)
				assert.NoError(t, err)
				if assert.NotNil(t, col) {
					assert.Equal(t, keyName, col.KeyName)
				}
			}()
		}
	}
	wg.Wait()

	assert.Equal(t, int64(1), api.listRequests.Load(), "expected a single list request")
	assert.Zero(t, api.keyNameRequests.Load(), "expected no per-key-name requests")

	// List is served from the same cache
	columns, err := c.Columns.List(ctx, "test-dataset")
	require.NoError(t, err)
	assert.Len(t, columns, 2)
	assert.Equal(t, int64(1), api.listRequests.Load())
}

func TestColumns_CachingIsDisabledByDefault(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	api := &columnTestAPI{
		columns: map[string][]client.Column{
			"test-dataset": {{ID: "id-1", KeyName: "column_1"}},
		},
	}
	// no withReadCaching: the default must preserve today's behaviour
	c := newColumnTestClient(t, api)

	for range 3 {
		_, err := c.Columns.GetByKeyName(ctx, "test-dataset", "column_1")
		require.NoError(t, err)
	}

	assert.Equal(t, int64(3), api.keyNameRequests.Load(), "expected one direct lookup per read")
	assert.Zero(t, api.listRequests.Load(), "expected no list requests when caching is off")
}

func TestColumns_MissingKeyNameFallsBackToDirectLookup(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	api := &columnTestAPI{columns: map[string][]client.Column{"test-dataset": {}}}
	c := newColumnTestClient(t, api, withReadCaching)

	_, err := c.Columns.GetByKeyName(ctx, "test-dataset", "column_gone")

	var detailedErr client.DetailedError
	require.ErrorAs(t, err, &detailedErr)
	assert.True(t, detailedErr.IsNotFound(), "expected NotFound to propagate through the cache")
	assert.Equal(t, int64(1), api.keyNameRequests.Load(), "expected the miss to be confirmed with a direct lookup")
}

func TestColumns_WritesInvalidateTheCache(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	api := &columnTestAPI{columns: map[string][]client.Column{"test-dataset": {}}}
	c := newColumnTestClient(t, api, withReadCaching)

	// prime the cache with the empty dataset
	columns, err := c.Columns.List(ctx, "test-dataset")
	require.NoError(t, err)
	assert.Empty(t, columns)

	created, err := c.Columns.Create(ctx, "test-dataset", &client.Column{KeyName: "column_new"})
	require.NoError(t, err)

	// the create invalidated the cached list, so the new column is
	// visible without a per-key-name lookup
	col, err := c.Columns.GetByKeyName(ctx, "test-dataset", "column_new")
	require.NoError(t, err)
	assert.Equal(t, "column_new", col.KeyName)
	assert.Equal(t, int64(2), api.listRequests.Load(), "expected the list to be refetched after the write")
	assert.Zero(t, api.keyNameRequests.Load())

	// so does an update
	created.Description = "updated"
	_, err = c.Columns.Update(ctx, "test-dataset", created)
	require.NoError(t, err)
	col, err = c.Columns.GetByKeyName(ctx, "test-dataset", "column_new")
	require.NoError(t, err)
	assert.Equal(t, "updated", col.Description)
	assert.Equal(t, int64(3), api.listRequests.Load())

	// and a delete
	require.NoError(t, c.Columns.Delete(ctx, "test-dataset", created.ID))
	_, err = c.Columns.GetByKeyName(ctx, "test-dataset", "column_new")
	require.Error(t, err, "expected the deleted column to be gone from the cache")
	assert.Equal(t, int64(4), api.listRequests.Load())
}

func TestColumns_CacheIsPerDataset(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	api := &columnTestAPI{
		columns: map[string][]client.Column{
			"dataset-a": {{ID: "id-1", KeyName: "column_1"}},
			"dataset-b": {{ID: "id-2", KeyName: "column_1"}},
		},
	}
	c := newColumnTestClient(t, api, withReadCaching)

	a, err := c.Columns.GetByKeyName(ctx, "dataset-a", "column_1")
	require.NoError(t, err)
	b, err := c.Columns.GetByKeyName(ctx, "dataset-b", "column_1")
	require.NoError(t, err)

	assert.Equal(t, "id-1", a.ID)
	assert.Equal(t, "id-2", b.ID, "each dataset must be cached under its own key")
	assert.Equal(t, int64(2), api.listRequests.Load())

	// a write to one dataset does not invalidate the other
	_, err = c.Columns.Create(ctx, "dataset-a", &client.Column{KeyName: "column_2"})
	require.NoError(t, err)
	_, err = c.Columns.GetByKeyName(ctx, "dataset-b", "column_1")
	require.NoError(t, err)
	assert.Equal(t, int64(2), api.listRequests.Load(), "dataset-b's cache should be untouched")
}
