package validation_test

import (
	"context"
	"fmt"
	"strings"
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
		val              types.String
		expectError      bool
		expectedErrorMsg string
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
			val: types.StringValue(`{"calculations": [{"op": "AVG", "column": "duration_ms"}], "filters": [{"column": "status", "op": "!=", "value": "error"}], "breakdowns": ["service"]}`),
		},
		"valid formula with named calculations": {
			val: types.StringValue(`{
				"calculations": [
					{"op": "COUNT", "name": "total"},
					{"op": "COUNT", "name": "errors", "filters": [{"column": "status", "op": "=", "value": "error"}]}
				],
				"formulas": [{"name": "error_rate", "expression": "DIV($errors, $total)"}]
			}`),
		},
		"valid formula with having": {
			val: types.StringValue(`{
				"calculations": [
					{"op": "COUNT", "name": "total"},
					{"op": "COUNT", "name": "errors", "filters": [{"column": "status", "op": "=", "value": "error"}]},
					{"op": "SUM", "column": "bytes", "name": "total_bytes"}
				],
				"formulas": [{"name": "error_rate", "expression": "DIV($errors, $total)"}],
				"havings": [{"column": "error_rate", "op": ">", "value": 1000}]
			}`),
		},

		// --- Invalid cases ---
		"invalid json": {
			val:              types.StringValue("whoop"),
			expectError:      true,
			expectedErrorMsg: "value must be a valid Trigger Query Specification",
		},
		"invalid HEATMAP calculation": {
			val:              types.StringValue(`{"calculations": [{"op": "HEATMAP", "column": "duration_ms"}]}`),
			expectError:      true,
			expectedErrorMsg: "Trigger queries cannot use HEATMAP calculations.",
		},
		"invalid CONCURRENCY calculation": {
			val:              types.StringValue(`{"calculations": [{"op": "CONCURRENCY"}]}`),
			expectError:      true,
			expectedErrorMsg: "Trigger queries cannot use CONCURRENCY calculations.",
		},
		"invalid multiple calculations without formula": {
			val:              types.StringValue(`{"calculations": [{"op": "COUNT"}, {"op": "AVG", "column": "duration_ms"}]}`),
			expectError:      true,
			expectedErrorMsg: "Trigger queries must contain a single calculation, but found COUNT, AVG(duration_ms)",
		},
		"invalid orders": {
			val:              types.StringValue(`{"calculations": [{"op": "COUNT"}], "orders": [{"op": "COUNT", "order": "descending"}]}`),
			expectError:      true,
			expectedErrorMsg: "Trigger queries cannot use orders.",
		},
		"invalid limit": {
			val:              types.StringValue(`{"calculations": [{"op": "COUNT"}], "limit": 10}`),
			expectError:      true,
			expectedErrorMsg: "Trigger queries cannot use limit.",
		},
		"invalid start_time": {
			val:              types.StringValue(`{"calculations": [{"op": "COUNT"}], "start_time": 1454808600}`),
			expectError:      true,
			expectedErrorMsg: "Trigger queries cannot use start_time or end_time.",
		},
		"invalid end_time": {
			val:              types.StringValue(`{"calculations": [{"op": "COUNT"}], "end_time": 1454808600}`),
			expectError:      true,
			expectedErrorMsg: "Trigger queries cannot use start_time or end_time.",
		},
		"valid granularity": {
			val: types.StringValue(`{"calculations": [{"op": "COUNT"}], "granularity": 120}`),
		},
		"valid RATE_AVG with granularity": {
			val: types.StringValue(`{"calculations": [{"op": "RATE_AVG", "column": "http.server.requests"}], "granularity": 300}`),
		},
		"valid calculated field with granularity": {
			val: types.StringValue(`{"calculations": [{"op": "AVG", "column": "request_rate_5m"}], "calculated_fields": [{"name": "request_rate_5m", "expression": "RATE($http.server.requests, 300)"}], "granularity": 60}`),
		},
		"invalid more than 1 having": {
			val: types.StringValue(`{
				"calculations": [{"op": "COUNT"}, {"op": "AVG", "column": "duration_ms"}, {"op": "MAX", "column": "duration_ms"}],
				"havings": [
					{"calculate_op": "COUNT", "op": ">", "value": 5},
					{"calculate_op": "AVG", "column": "duration_ms", "op": ">", "value": 100}
				]
			}`),
			expectError:      true,
			expectedErrorMsg: "Trigger queries support at most 1 having clause, but found 2.",
		},
		"invalid more than 1 formula": {
			val: types.StringValue(`{
				"calculations": [
					{"op": "COUNT", "name": "a"},
					{"op": "AVG", "column": "duration_ms", "name": "b"}
				],
				"formulas": [
					{"name": "f1", "expression": "DIV($a, $b)"},
					{"name": "f2", "expression": "MUL($a, $b)"}
				]
			}`),
			expectError:      true,
			expectedErrorMsg: "Trigger queries support at most 1 formula, but found 2.",
		},
		"invalid named calculation without formula": {
			val: types.StringValue(`{
				"calculations": [{"op": "COUNT", "name": "total"}]
			}`),
			expectError:      true,
			expectedErrorMsg: "Trigger queries without formulas cannot use named calculations or calculation-level filters.",
		},
		"invalid calculation-level filters without formula": {
			val: types.StringValue(`{
				"calculations": [{"op": "COUNT", "filters": [{"column": "status", "op": "=", "value": "error"}]}]
			}`),
			expectError:      true,
			expectedErrorMsg: "Trigger queries without formulas cannot use named calculations or calculation-level filters.",
		},
		"invalid global filters with named aggregates": {
			val: types.StringValue(`{
				"calculations": [
					{"op": "COUNT", "name": "total"},
					{"op": "COUNT", "name": "errors", "filters": [{"column": "status", "op": "=", "value": "error"}]}
				],
				"formulas": [{"name": "error_rate", "expression": "DIV($errors, $total)"}],
				"filters": [{"column": "service", "op": "=", "value": "web"}]
			}`),
			expectError:      true,
			expectedErrorMsg: "Trigger queries cannot use global filters when calculations have names or aggregate filters. Use calculation-level filters instead.",
		},
		"invalid global filters with aggregate filters": {
			val: types.StringValue(`{
				"calculations": [
					{"op": "COUNT", "name": "errors", "filters": [{"column": "status", "op": "=", "value": "error"}]}
				],
				"formulas": [{"name": "error_count", "expression": "$errors"}],
				"filters": [{"column": "service", "op": "=", "value": "web"}]
			}`),
			expectError:      true,
			expectedErrorMsg: "Trigger queries cannot use global filters when calculations have names or aggregate filters. Use calculation-level filters instead.",
		},
		"invalid duplicate calculation names": {
			val: types.StringValue(`{
				"calculations": [
					{"op": "COUNT", "name": "total"},
					{"op": "AVG", "column": "duration_ms", "name": "total"}
				],
				"formulas": [{"name": "f", "expression": "$total"}]
			}`),
			expectError:      true,
			expectedErrorMsg: "Trigger queries cannot have duplicate calculation names.",
		},
		"invalid calculation name conflicts with formula name": {
			val: types.StringValue(`{
				"calculations": [
					{"op": "COUNT", "name": "result"},
					{"op": "AVG", "column": "duration_ms", "name": "b"}
				],
				"formulas": [{"name": "result", "expression": "$b"}]
			}`),
			expectError:      true,
			expectedErrorMsg: `Trigger queries cannot have a formula with the same name as a calculation: "result".`,
		},
		"invalid multiple calculations not obscured by having": {
			val: types.StringValue(`{
				"calculations": [{"op": "COUNT"}, {"op": "AVG", "column": "duration_ms"}, {"op": "P99", "column": "duration_ms"}],
				"havings": [{"calculate_op": "P99", "column": "duration_ms", "op": ">", "value": 1000}]
			}`),
			expectError:      true,
			expectedErrorMsg: "Trigger queries must contain a single calculation, but found COUNT, AVG(duration_ms)",
		},
	}

	// Add test for >100 calculations with formula
	calcs := make([]string, 101)
	for i := range calcs {
		calcs[i] = fmt.Sprintf(`{"op": "COUNT", "name": "c%d"}`, i)
	}
	tests["invalid more than 100 calculations with formula"] = testCase{
		val: types.StringValue(fmt.Sprintf(`{
			"calculations": [%s],
			"formulas": [{"name": "f", "expression": "$c0"}]
		}`, strings.Join(calcs, ","))),
		expectError:      true,
		expectedErrorMsg: "Trigger queries with formulas support at most 100 calculations, but found 101.",
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

			if test.expectedErrorMsg != "" {
				assert.True(t, response.Diagnostics.HasError(), "expected an error but got none")
				if response.Diagnostics.HasError() {
					found := false
					for _, diag := range response.Diagnostics.Errors() {
						if strings.Contains(diag.Detail(), test.expectedErrorMsg) {
							found = true
							break
						}
					}
					assert.True(t, found, "expected error message %q but got: %s", test.expectedErrorMsg, response.Diagnostics)
				}
			}
		})
	}
}
