package honeycombio

import (
	"context"
	"fmt"
	"testing"
	"time"

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

	b := &Board{
		Name: fmt.Sprintf("Test QuerySpec, created at %v", time.Now()),

		Queries: []BoardQuery{
			{
				Dataset: dataset,
				Query:   &query,
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
	assert.Equal(t, &query, b.Queries[0].Query)
}

func TestCalcuationOps(t *testing.T) {
	ctx := context.Background()
	c := newTestClient(t)
	dataset := testDataset(t)

	b := &Board{
		Name: "go-honeycombio: TestCalcula§tionOps",
	}
	b, err := c.Boards.Create(ctx, b)
	assert.NoError(t, err)

	defer c.Boards.Delete(ctx, b.ID)

	for _, calculationOp := range CalculationOps() {
		column := StringPtr("duration_ms")
		if calculationOp == CalculationOpCount {
			column = nil
		}

		q := QuerySpec{
			Calculations: []CalculationSpec{
				{
					Op:     calculationOp,
					Column: column,
				},
			},
		}
		b.Queries = []BoardQuery{{Dataset: dataset, Query: &q}}

		_, err = c.Boards.Update(ctx, b)
		assert.NoError(t, err, fmt.Sprintf("Failed to create board that contains calcuation with op \"%v\"", calculationOp))
	}
}

func TestFilterOps(t *testing.T) {
	ctx := context.Background()
	c := newTestClient(t)
	dataset := testDataset(t)

	b := &Board{
		Name: "go-honeycombio: TestFilterOps",
	}
	b, err := c.Boards.Create(ctx, b)
	assert.NoError(t, err)

	defer c.Boards.Delete(ctx, b.ID)

	for _, filterOp := range FilterOps() {
		var value interface{}

		switch filterOp {
		case FilterOpExists, FilterOpDoesNotExist:
			value = nil
		case FilterOpIn, FilterOpNotIn:
			value = []string{"foo", "bar"}
		default:
			value = "foo"
		}

		q := QuerySpec{
			Filters: []FilterSpec{
				{
					Column: "column_1",
					Op:     filterOp,
					Value:  value,
				},
			},
		}
		b.Queries = []BoardQuery{{Dataset: dataset, Query: &q}}

		_, err = c.Boards.Update(ctx, b)
		assert.NoError(t, err, fmt.Sprintf("Failed to create board that contains filter with op \"%v\"", filterOp))
	}
}
