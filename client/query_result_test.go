package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueryResults(t *testing.T) {
	ctx := context.Background()

	var result *QueryResult
	var err error

	c := newTestClient(t)
	dataset := testDataset(t)

	query, err := c.Queries.Create(ctx, dataset, &QuerySpec{
		Calculations: []CalculationSpec{
			{
				Op: "COUNT",
			},
		},
		TimeRange: ToPtr(60 * 60 * 24),
	})
	require.NoError(t, err)

	t.Run("Create", func(t *testing.T) {
		result, err = c.QueryResults.Create(ctx, dataset, &QueryResultRequest{
			ID: *query.ID,
		})

		require.NoError(t, err)
		assert.NotEmpty(t, result.ID, "result missing ID")
	})

	t.Run("Get", func(t *testing.T) {
		err := c.QueryResults.Get(ctx, dataset, result)

		assert.NoError(t, err)
		assert.True(t, result.Complete, "query result didn't complete")
		assert.NotNil(t, result.Data.Series, "empty data series")
		assert.NotNil(t, result.Data.Results, "empty data results")
		assert.NotEmpty(t, result.Links.GraphUrl, "empty result graph")
		assert.NotEmpty(t, result.Links.Url, "empty result Url")
	})

	t.Run("Fail to Get bogus Query Result", func(t *testing.T) {
		err := c.QueryResults.Get(ctx, dataset, &QueryResult{ID: "abcd1234"})

		var de DetailedError
		assert.Error(t, err)
		assert.ErrorAs(t, err, &de)
		assert.True(t, de.IsNotFound())
	})
}
