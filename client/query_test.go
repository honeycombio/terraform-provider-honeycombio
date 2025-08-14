package client_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
)

func TestQueries(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	c := newTestClient(t)
	dataset := testDataset(t)

	var query *client.QuerySpec
	var err error

	floatCol, col1, col2 := createRandomTestColumns(t, c, dataset)

	t.Run("Create", func(t *testing.T) {
		data := &client.QuerySpec{
			Calculations: []client.CalculationSpec{
				{
					Op: client.CalculationOpCount,
				},
				{
					Op:     client.CalculationOpHeatmap,
					Column: &floatCol.KeyName,
				},
			},
			Filters: []client.FilterSpec{
				{
					Column: col1.KeyName,
					Op:     client.FilterOpExists,
				},
				{
					Column: floatCol.KeyName,
					Op:     client.FilterOpSmallerThan,
					Value:  10000.0,
				},
			},
			FilterCombination: client.FilterCombinationOr,
			Breakdowns:        []string{col1.KeyName, col2.KeyName},
			Orders: []client.OrderSpec{
				{
					Column: &col1.KeyName,
				},
				{
					Op:    client.ToPtr(client.CalculationOpCount),
					Order: client.ToPtr(client.SortOrderDesc),
				},
			},
			Limit:       client.ToPtr(100),
			TimeRange:   client.ToPtr(3600), // 1 hour
			Granularity: client.ToPtr(60),   // 1 minute
		}

		query, err = c.Queries.Create(ctx, dataset, data)
		require.NoError(t, err)

		data.ID = query.ID
		assert.Equal(t, data, query)
	})

	t.Run("Get", func(t *testing.T) {
		q, err := c.Queries.Get(ctx, dataset, *query.ID)
		require.NoError(t, err)

		assert.Equal(t, query, q)
	})
}
