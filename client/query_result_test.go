package client_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
)

func TestQueryResults(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	var result *client.QueryResult
	var err error

	c := newTestClient(t)
	dataset := testDataset(t)

	query, err := c.Queries.Create(ctx, dataset, &client.QuerySpec{
		Calculations: []client.CalculationSpec{
			{
				Op: "COUNT",
			},
		},
		TimeRange: client.ToPtr(60 * 60 * 24),
	})
	require.NoError(t, err)

	t.Run("Create", func(t *testing.T) {
		result, err = c.QueryResults.Create(ctx, dataset, &client.QueryResultRequest{
			ID: *query.ID,
		})

		require.NoError(t, err)
		assert.NotEmpty(t, result.ID, "result missing ID")
	})

	t.Run("Get", func(t *testing.T) {
		err := c.QueryResults.Get(ctx, dataset, result)

		require.NoError(t, err)
		assert.True(t, result.Complete, "query result didn't complete")
		assert.NotEmpty(t, result.Data.Results, "results should not be empty")
		assert.NotEmpty(t, result.Links.GraphUrl, "empty result graph")
		assert.NotEmpty(t, result.Links.Url, "empty result Url")
		assert.Empty(t, result.Data.Series, "data series should be empty")
	})

	t.Run("Fail to Get bogus Query Result", func(t *testing.T) {
		err := c.QueryResults.Get(ctx, dataset, &client.QueryResult{ID: "abcd1234"})

		var de client.DetailedError
		require.Error(t, err)
		require.ErrorAs(t, err, &de)
		assert.True(t, de.IsNotFound())
	})
}
