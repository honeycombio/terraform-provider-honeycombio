package honeycombio

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// create a board with an elaborate QuerySpec as smoke test
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
		FilterCombination: FilterCombinationPtr(FilterCombinationOr),
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
		Limit: IntPtr(100),
	}

	b := &Board{
		Name: "go-honeycombio - test QuerySpec",

		Queries: []BoardQuery{
			{
				Dataset: dataset,
				Query:   query,
			},
		},
	}

	b, err := c.Boards.Create(ctx, b)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := c.Boards.Delete(ctx, b.ID)
		if err != nil {
			t.Fatal(err)
		}
	}()

	assert.Len(t, b.Queries, 1)
	assert.Equal(t, query, b.Queries[0].Query)
}
