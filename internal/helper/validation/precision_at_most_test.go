package validation_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/validation"
)

func TestValidation_PrecisionAtMost(t *testing.T) {
	t.Parallel()

	type testCase struct {
		value        types.Float64
		maxPrecision int64
		expectError  bool
	}

	tests := map[string]testCase{
		"unknown value": {
			value:        types.Float64Unknown(),
			maxPrecision: 5,
		},
		"null value": {
			value:        types.Float64Null(),
			maxPrecision: 4,
		},
		"valid value at precision limit": {
			value:        types.Float64Value(0.001),
			maxPrecision: 3,
		},
		"valid value under precision limit": {
			value:        types.Float64Value(5),
			maxPrecision: 2,
		},
		"valid large value under precision limit": {
			value:        types.Float64Value(123),
			maxPrecision: 2,
		},
		// trailing zeros won't impact the conversion from percent to ppm so there's no need to fail
		"valid value over precision limit with trailing zeros doesn't fail because the trailing zeros don't matter": {
			value:        types.Float64Value(99.00000),
			maxPrecision: 4,
		},
		"invalid value over precision limit": {
			value:        types.Float64Value(99.99999),
			maxPrecision: 4,
			expectError:  true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			request := validator.Float64Request{
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				ConfigValue:    test.value,
			}
			response := validator.Float64Response{}
			validation.Float64PrecisionAtMost(test.maxPrecision).ValidateFloat64(context.Background(), request, &response)

			if !response.Diagnostics.HasError() && test.expectError {
				t.Fatal("expected error, got no error")
			}

			if response.Diagnostics.HasError() && !test.expectError {
				t.Fatalf("got unexpected error: %s", response.Diagnostics)
			}
		})
	}
}
