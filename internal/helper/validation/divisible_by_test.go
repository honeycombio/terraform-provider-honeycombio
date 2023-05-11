package validation_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/validation"
)

func Test_BetweenValidator(t *testing.T) {
	t.Parallel()

	type testCase struct {
		val         types.Int64
		divisor     int64
		expectError bool
	}
	tests := map[string]testCase{
		"unknown Int64": {
			val:     types.Int64Unknown(),
			divisor: 3,
		},
		"null Int64": {
			val:     types.Int64Null(),
			divisor: 3,
		},
		"valid integer as Int64": {
			val:     types.Int64Value(3600),
			divisor: 60,
		},
		"valid negative integer as Int64": {
			val:     types.Int64Value(-6),
			divisor: 2,
		},
		"not divisble by integer as Int64": {
			val:         types.Int64Value(20),
			divisor:     3,
			expectError: true,
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			request := validator.Int64Request{
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				ConfigValue:    test.val,
			}
			response := validator.Int64Response{}
			validation.Int64DivisibleBy(test.divisor).ValidateInt64(context.TODO(), request, &response)

			if !response.Diagnostics.HasError() && test.expectError {
				t.Fatal("expected error, got no error")
			}

			if response.Diagnostics.HasError() && !test.expectError {
				t.Fatalf("got unexpected error: %s", response.Diagnostics)
			}
		})
	}
}
