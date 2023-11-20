package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueries(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	c := newTestClient(t)
	dataset := testDataset(t)

	var query *QuerySpec
	var err error

	floatCol, col1, col2 := createRandomTestColumns(t, c, dataset)

	t.Run("Create", func(t *testing.T) {
		data := &QuerySpec{
			Calculations: []CalculationSpec{
				{
					Op: CalculationOpCount,
				},
				{
					Op:     CalculationOpHeatmap,
					Column: &floatCol.KeyName,
				},
			},
			Filters: []FilterSpec{
				{
					Column: col1.KeyName,
					Op:     FilterOpExists,
				},
				{
					Column: floatCol.KeyName,
					Op:     FilterOpSmallerThan,
					Value:  10000.0,
				},
			},
			FilterCombination: FilterCombinationOr,
			Breakdowns:        []string{col1.KeyName, col2.KeyName},
			Orders: []OrderSpec{
				{
					Column: &col1.KeyName,
				},
				{
					Op:    ToPtr(CalculationOpCount),
					Order: ToPtr(SortOrderDesc),
				},
			},
			Limit:       ToPtr(100),
			TimeRange:   ToPtr(3600), // 1 hour
			Granularity: ToPtr(60),   // 1 minute
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
