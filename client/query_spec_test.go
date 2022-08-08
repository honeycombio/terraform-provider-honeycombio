package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// create a query with an elaborate QuerySpec as smoke test
func TestQuerySpec(t *testing.T) {
	ctx := context.Background()

	c := newTestClient(t)
	dataset := testDataset(t)

	query := QuerySpec{
		Calculations: []CalculationSpec{
			{
				Op: CalculationOpCount,
			},
			{
				Op:     CalculationOpHeatmap,
				Column: StringPtr("duration_ms"),
			},
			{
				Op:     CalculationOpP99,
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
			{
				Column: "column_1",
				Op:     FilterOpNotEquals,
				Value:  "",
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
		Havings: []HavingSpec{
			{
				Column:      StringPtr("duration_ms"),
				Op:          HavingOpPtr(HavingOpGreaterThan),
				CalculateOp: CalculationOpPtr(CalculationOpP99),
				Value:       1000.0,
			},
		},
		Limit:       IntPtr(100),
		TimeRange:   IntPtr(3600), // 1 hour
		Granularity: IntPtr(60),   // 1 minute
	}

	_, err := c.Queries.Create(ctx, dataset, &query)
	assert.NoError(t, err)
}
