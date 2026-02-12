package validation_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"

	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/validation"
)

func Test_TriggerQuerySpecValidator(t *testing.T) {
	t.Parallel()

	type testCase struct {
		val         types.String
		expectError bool
	}
	tests := map[string]testCase{
		// --- Valid cases ---
		"unknown": {
			val: types.StringUnknown(),
		},
		"null": {
			val: types.StringNull(),
		},
		"valid single calculation": {
			val: types.StringValue(`{"calculations": [{"op": "COUNT"}]}`),
		},
		"valid single calculation with having match": {
			val: types.StringValue(`{"calculations": [{"op": "COUNT"}], "havings": [{"calculate_op": "COUNT", "op": ">", "value": 5}]}`),
		},
		"valid two calcs one matches having": {
			val: types.StringValue(`{"calculations": [{"op": "COUNT"}, {"op": "P99", "column": "duration_ms"}], "havings": [{"calculate_op": "COUNT", "op": ">", "value": 5}]}`),
		},
		"valid single calculation with filters and breakdowns": {
			val: types.StringValue(`{"calculations": [{"op": "AVG", "column": "duration_ms"}], "filters": [{"column": "status", "op": "=", "value": "error"}], "breakdowns": ["service"]}`),
		},

		// --- Invalid cases ---
		"invalid json": {
			val:         types.StringValue("whoop"),
			expectError: true,
		},
		"invalid HEATMAP calculation": {
			val:         types.StringValue(`{"calculations": [{"op": "HEATMAP", "column": "duration_ms"}]}`),
			expectError: true,
		},
		"invalid CONCURRENCY calculation": {
			val:         types.StringValue(`{"calculations": [{"op": "CONCURRENCY"}]}`),
			expectError: true,
		},
		"invalid multiple calculations without formula": {
			val:         types.StringValue(`{"calculations": [{"op": "COUNT"}, {"op": "AVG", "column": "duration_ms"}]}`),
			expectError: true,
		},
		"invalid orders": {
			val:         types.StringValue(`{"calculations": [{"op": "COUNT"}], "orders": [{"op": "COUNT", "order": "descending"}]}`),
			expectError: true,
		},
		"invalid limit": {
			val:         types.StringValue(`{"calculations": [{"op": "COUNT"}], "limit": 10}`),
			expectError: true,
		},
		"invalid start_time": {
			val:         types.StringValue(`{"calculations": [{"op": "COUNT"}], "start_time": 1454808600}`),
			expectError: true,
		},
		"invalid end_time": {
			val:         types.StringValue(`{"calculations": [{"op": "COUNT"}], "end_time": 1454808600}`),
			expectError: true,
		},
		"invalid granularity": {
			val:         types.StringValue(`{"calculations": [{"op": "COUNT"}], "granularity": 120}`),
			expectError: true,
		},
		"invalid multiple calculations not obscured by having": {
			val: types.StringValue(`{
				"calculations": [{"op": "COUNT"}, {"op": "AVG", "column": "duration_ms"}, {"op": "P99", "column": "duration_ms"}],
				"havings": [{"calculate_op": "P99", "column": "duration_ms", "op": ">", "value": 1000}]
			}`),
			expectError: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			request := validator.StringRequest{
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				ConfigValue:    test.val,
			}
			response := validator.StringResponse{}
			validation.ValidTriggerQuerySpec().ValidateString(context.Background(), request, &response)

			assert.Equal(t,
				test.expectError,
				response.Diagnostics.HasError(),
				"unexpected result for %q: %s", name, response.Diagnostics,
			)
		})
	}
}
