package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// create a query with an elaborate QuerySpec as smoke test
func TestQuerySpec(t *testing.T) {
	t.Parallel()

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
				Column: ToPtr("duration_ms"),
			},
			{
				Op:     CalculationOpP99,
				Column: ToPtr("duration_ms"),
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
				Column: ToPtr("column_1"),
			},
			{
				Op:    ToPtr(CalculationOpCount),
				Order: ToPtr(SortOrderDesc),
			},
		},
		Havings: []HavingSpec{
			{
				Column:      ToPtr("duration_ms"),
				Op:          ToPtr(HavingOpGreaterThan),
				CalculateOp: ToPtr(CalculationOpP99),
				Value:       1000.0,
			},
		},
		Limit:       ToPtr(100),
		TimeRange:   ToPtr(3600), // 1 hour
		Granularity: ToPtr(60),   // 1 minute
	}

	_, err := c.Queries.Create(ctx, dataset, &query)
	assert.NoError(t, err)
}

func TestQuerySpec_EquivalentTo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		a, b QuerySpec
		want bool
	}{
		{"Empty", QuerySpec{}, QuerySpec{}, true},
		{
			"Empty Defaults",
			QuerySpec{
				Calculations: []CalculationSpec{
					{
						Op: "COUNT",
					},
				},
				FilterCombination: "AND",
				TimeRange:         ToPtr(DefaultQueryTimeRange),
				// Granularity may be exported out of the Query Builder as '0' when not provided
				Granularity: ToPtr(0),
			},
			QuerySpec{},
			true,
		},
		{
			"Equivalent but shuffled",
			QuerySpec{
				Calculations: []CalculationSpec{
					{
						Op:     "HEATMAP",
						Column: ToPtr("duration_ms"),
					},
					{
						Op: "COUNT",
					},
				},
				Filters: []FilterSpec{
					{
						Column: "colA",
						Op:     "=",
						Value:  "a",
					},
					{
						Column: "colC",
						Op:     "=",
						Value:  "c",
					},
					{
						Column: "colB",
						Op:     "=",
						Value:  "b",
					},
				},
				Breakdowns: []string{"colB", "colA"},
				TimeRange:  ToPtr(DefaultQueryTimeRange),
			},
			QuerySpec{
				Calculations: []CalculationSpec{
					{
						Op:     "HEATMAP",
						Column: ToPtr("duration_ms"),
					},
					{
						Op: "COUNT",
					},
				},
				Filters: []FilterSpec{
					{
						Column: "colC",
						Op:     "=",
						Value:  "c",
					},
					{
						Column: "colB",
						Op:     "=",
						Value:  "b",
					},
					{
						Column: "colA",
						Op:     "=",
						Value:  "a",
					},
				},
				Breakdowns:        []string{"colB", "colA"},
				FilterCombination: "AND",
			},
			true,
		},
		{
			"Calculation order matters",
			QuerySpec{
				Calculations: []CalculationSpec{
					{
						Op:     "HEATMAP",
						Column: ToPtr("duration_ms"),
					},
					{
						Op: "COUNT",
					},
				},
			},
			QuerySpec{
				Calculations: []CalculationSpec{
					{
						Op: "COUNT",
					},
					{
						Op:     "HEATMAP",
						Column: ToPtr("duration_ms"),
					},
				},
			},
			false,
		},
		{
			"Different time ranges",
			QuerySpec{
				Calculations: []CalculationSpec{
					{
						Op: "COUNT",
					},
				},
				TimeRange: ToPtr(1800),
			},
			QuerySpec{
				Calculations: []CalculationSpec{
					{
						Op: "COUNT",
					},
				},
			},
			false,
		},
		{
			"Different FilterCombinations",
			QuerySpec{
				FilterCombination: "OR",
			},
			QuerySpec{},
			false,
		},
		{
			"Calculation different from DefaultCalc",
			QuerySpec{
				Calculations: []CalculationSpec{
					{
						Op:     "MIN",
						Column: ToPtr("metrics.cpu.utilization"),
					},
				},
			},
			QuerySpec{},
			false,
		},
		{
			"Different Calculations",
			QuerySpec{
				Calculations: []CalculationSpec{
					{
						Op:     "MIN",
						Column: ToPtr("metrics.cpu.utilization"),
					},
				},
			},
			QuerySpec{
				Calculations: []CalculationSpec{
					{
						Op:     "MAX",
						Column: ToPtr("metrics.cpu.utilization"),
					},
				},
			},
			false,
		},
		{
			"Different Number of Calculations",
			QuerySpec{
				Calculations: []CalculationSpec{
					{
						Op:     "MIN",
						Column: ToPtr("metrics.cpu.utilization"),
					},
				},
			},
			QuerySpec{
				Calculations: []CalculationSpec{
					{
						Op: "COUNT_DISTINCT",
					},
					{
						Op:     "MAX",
						Column: ToPtr("metrics.cpu.utilization"),
					},
				},
			},
			false,
		},
		{
			"Equivalent column orders",
			QuerySpec{
				Orders: []OrderSpec{
					{Column: ToPtr("column_1")},
				},
			},
			QuerySpec{
				Orders: []OrderSpec{
					{
						Column: ToPtr("column_1"),
						Order:  ToPtr(SortOrderAsc),
					},
				},
			},
			true,
		},
		{
			"Not equivalent column orders",
			QuerySpec{
				Orders: []OrderSpec{
					{Column: ToPtr("column_2")},
				},
			},
			QuerySpec{
				Orders: []OrderSpec{
					{
						Column: ToPtr("column_1"),
						Order:  ToPtr(SortOrderAsc),
					},
				},
			},
			false,
		},
		{
			"Equivalent Op orders with unspecified default",
			QuerySpec{
				Orders: []OrderSpec{
					{
						Op:    ToPtr(CalculationOpCount),
						Order: ToPtr(SortOrderAsc),
					},
				},
			},
			QuerySpec{
				Orders: []OrderSpec{
					{Op: ToPtr(CalculationOpCount)},
				},
			},
			true,
		},
		{
			"Not equivalent Op orders",
			QuerySpec{
				Orders: []OrderSpec{
					{
						Op:     ToPtr(CalculationOpCountDistinct),
						Column: ToPtr("column_1"),
						Order:  ToPtr(SortOrderAsc),
					},
				},
			},
			QuerySpec{
				Orders: []OrderSpec{
					{
						Op:    ToPtr(CalculationOpCount),
						Order: ToPtr(SortOrderDesc),
					},
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.a.EquivalentTo(tt.b))
		})
	}
}
