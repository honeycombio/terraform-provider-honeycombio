package client

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
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
		TimeRange: IntPtr(60 * 60 * 24),
	})

	t.Run("Create", func(t *testing.T) {
		result, err = c.QueryResults.Create(ctx, dataset, &QueryResultRequest{
			ID: *query.ID,
		})

		assert.Nil(t, err, fmt.Sprintf("result errored: %v", err))
		assert.NotEmpty(t, result.ID, "result missing ID")
	})

	t.Run("Get", func(t *testing.T) {
		err := c.QueryResults.Get(ctx, dataset, result)

		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, result.Complete, "query result didn't complete")
		assert.NotNil(t, result.Data.Series, "empty data series")
		assert.NotNil(t, result.Data.Results, "empty data results")
		assert.NotEmpty(t, result.Links.GraphUrl, "empty result graph")
		assert.NotEmpty(t, result.Links.Url, "empty result Url")
	})

	t.Run("Get_notFound", func(t *testing.T) {
		err := c.QueryResults.Get(ctx, dataset, &QueryResult{ID: "abcd1234"})

		assert.ErrorIs(t, err, ErrNotFound)
	})
}
