package client_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
)

// create a query with an elaborate QuerySpec as smoke test
func TestQuerySpec(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	c := newTestClient(t)
	dataset := testDataset(t)

	query := client.QuerySpec{
		Calculations: []client.CalculationSpec{
			{
				Op: client.CalculationOpCount,
			},
			{
				Op:     client.CalculationOpHeatmap,
				Column: client.ToPtr("duration_ms"),
			},
			{
				Op:     client.CalculationOpP99,
				Column: client.ToPtr("duration_ms"),
			},
		},
		Filters: []client.FilterSpec{
			{
				Column: "column_1",
				Op:     client.FilterOpExists,
			},
			{
				Column: "duration_ms",
				Op:     client.FilterOpSmallerThan,
				Value:  10000.0,
			},
			{
				Column: "column_1",
				Op:     client.FilterOpNotEquals,
				Value:  "",
			},
		},
		FilterCombination: client.FilterCombinationOr,
		Breakdowns:        []string{"column_1", "column_2"},
		Orders: []client.OrderSpec{
			{
				Column: client.ToPtr("column_1"),
			},
			{
				Op:    client.ToPtr(client.CalculationOpCount),
				Order: client.ToPtr(client.SortOrderDesc),
			},
		},
		Havings: []client.HavingSpec{
			{
				Column:      client.ToPtr("duration_ms"),
				Op:          client.ToPtr(client.HavingOpGreaterThan),
				CalculateOp: client.ToPtr(client.CalculationOpP99),
				Value:       1000.0,
			},
		},
		Limit:       client.ToPtr(100),
		TimeRange:   client.ToPtr(3600), // 1 hour
		Granularity: client.ToPtr(60),   // 1 minute
	}

	_, err := c.Queries.Create(ctx, dataset, &query)
	assert.NoError(t, err)
}

func TestQuerySpec_EquivalentTo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		a, b client.QuerySpec
		want bool
	}{
		{"Empty", client.QuerySpec{}, client.QuerySpec{}, true},
		{
			"Empty Defaults",
			client.QuerySpec{
				Calculations: []client.CalculationSpec{
					{
						Op: "COUNT",
					},
				},
				FilterCombination: "AND",
				TimeRange:         client.ToPtr(client.DefaultQueryTimeRange),
				// Granularity may be exported out of the Query Builder as '0' when not provided
				Granularity: client.ToPtr(0),
				Breakdowns:  []string{},
				Orders:      []client.OrderSpec{},
			},
			client.QuerySpec{},
			true,
		},
		{
			"Equivalent but shuffled",
			client.QuerySpec{
				Calculations: []client.CalculationSpec{
					{
						Op:     "HEATMAP",
						Column: client.ToPtr("duration_ms"),
					},
					{
						Op: "COUNT",
					},
				},
				Filters: []client.FilterSpec{
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
				TimeRange:  client.ToPtr(client.DefaultQueryTimeRange),
			},
			client.QuerySpec{
				Calculations: []client.CalculationSpec{
					{
						Op:     "HEATMAP",
						Column: client.ToPtr("duration_ms"),
					},
					{
						Op: "COUNT",
					},
				},
				Filters: []client.FilterSpec{
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
			client.QuerySpec{
				Calculations: []client.CalculationSpec{
					{
						Op:     "HEATMAP",
						Column: client.ToPtr("duration_ms"),
					},
					{
						Op: "COUNT",
					},
				},
			},
			client.QuerySpec{
				Calculations: []client.CalculationSpec{
					{
						Op: "COUNT",
					},
					{
						Op:     "HEATMAP",
						Column: client.ToPtr("duration_ms"),
					},
				},
			},
			false,
		},
		{
			"Different time ranges",
			client.QuerySpec{
				Calculations: []client.CalculationSpec{
					{
						Op: "COUNT",
					},
				},
				TimeRange: client.ToPtr(1800),
			},
			client.QuerySpec{
				Calculations: []client.CalculationSpec{
					{
						Op: "COUNT",
					},
				},
			},
			false,
		},
		{
			"Different FilterCombinations",
			client.QuerySpec{
				FilterCombination: "OR",
			},
			client.QuerySpec{},
			false,
		},
		{
			"Calculation different from DefaultCalc",
			client.QuerySpec{
				Calculations: []client.CalculationSpec{
					{
						Op:     "MIN",
						Column: client.ToPtr("metrics.cpu.utilization"),
					},
				},
			},
			client.QuerySpec{},
			false,
		},
		{
			"Different Calculations",
			client.QuerySpec{
				Calculations: []client.CalculationSpec{
					{
						Op:     "MIN",
						Column: client.ToPtr("metrics.cpu.utilization"),
					},
				},
			},
			client.QuerySpec{
				Calculations: []client.CalculationSpec{
					{
						Op:     "MAX",
						Column: client.ToPtr("metrics.cpu.utilization"),
					},
				},
			},
			false,
		},
		{
			"Different Number of Calculations",
			client.QuerySpec{
				Calculations: []client.CalculationSpec{
					{
						Op:     "MIN",
						Column: client.ToPtr("metrics.cpu.utilization"),
					},
				},
			},
			client.QuerySpec{
				Calculations: []client.CalculationSpec{
					{
						Op: "COUNT_DISTINCT",
					},
					{
						Op:     "MAX",
						Column: client.ToPtr("metrics.cpu.utilization"),
					},
				},
			},
			false,
		},
		{
			"Equivalent column orders",
			client.QuerySpec{
				Orders: []client.OrderSpec{
					{Column: client.ToPtr("column_1")},
				},
			},
			client.QuerySpec{
				Orders: []client.OrderSpec{
					{
						Column: client.ToPtr("column_1"),
						Order:  client.ToPtr(client.SortOrderAsc),
					},
				},
			},
			true,
		},
		{
			"Not equivalent column orders",
			client.QuerySpec{
				Orders: []client.OrderSpec{
					{Column: client.ToPtr("column_2")},
				},
			},
			client.QuerySpec{
				Orders: []client.OrderSpec{
					{
						Column: client.ToPtr("column_1"),
						Order:  client.ToPtr(client.SortOrderAsc),
					},
				},
			},
			false,
		},
		{
			"Equivalent Op orders with unspecified default",
			client.QuerySpec{
				Orders: []client.OrderSpec{
					{
						Op:    client.ToPtr(client.CalculationOpCount),
						Order: client.ToPtr(client.SortOrderAsc),
					},
				},
			},
			client.QuerySpec{
				Orders: []client.OrderSpec{
					{Op: client.ToPtr(client.CalculationOpCount)},
				},
			},
			true,
		},
		{
			"Not equivalent Op orders",
			client.QuerySpec{
				Orders: []client.OrderSpec{
					{
						Op:     client.ToPtr(client.CalculationOpCountDistinct),
						Column: client.ToPtr("column_1"),
						Order:  client.ToPtr(client.SortOrderAsc),
					},
				},
			},
			client.QuerySpec{
				Orders: []client.OrderSpec{
					{
						Op:    client.ToPtr(client.CalculationOpCount),
						Order: client.ToPtr(client.SortOrderDesc),
					},
				},
			},
			false,
		},
		{
			"Not equivalent breakdowns",
			client.QuerySpec{
				Breakdowns: []string{"column_1"},
			},
			client.QuerySpec{},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.a.EquivalentTo(tt.b))
		})
	}
}
