package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueries(t *testing.T) {
	ctx := context.Background()

	c := newTestClient(t)
	dataset := testDataset(t)

	var query *QuerySpec
	var err error

	t.Run("Create", func(t *testing.T) {
		data := &QuerySpec{
			Calculations: []CalculationSpec{
				{
					Op: CalculationOpCount,
				},
				{
					Op:     CalculationOpHeatmap,
					Column: StringPtr("duration_ms"),
				},
			},
			Filters: []FilterSpec{
				{
					Column: "column_1",
					Op:     FilterOpExists,
				},
				{
					Column: "duration_ms",
					Op:     FilterOpSmallerThan,
					Value:  10000.0,
				},
			},
			FilterCombination: FilterCombinationOr,
			Breakdowns:        []string{"column_1", "column_2"},
			Orders: []OrderSpec{
				{
					Column: StringPtr("column_1"),
				},
				{
					Op:    CalculationOpPtr(CalculationOpCount),
					Order: SortOrderPtr(SortOrderDesc),
				},
			},
			Limit:       IntPtr(100),
			TimeRange:   IntPtr(3600), // 1 hour
			Granularity: IntPtr(60),   // 1 minute
		}

		query, err = c.Queries.Create(ctx, dataset, data)

		if err != nil {
			t.Fatal(err)
		}

		data.ID = query.ID
		assert.Equal(t, data, query)
	})

	t.Run("Get", func(t *testing.T) {
		q, err := c.Queries.Get(ctx, dataset, *query.ID)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, query, q)
	})
}
