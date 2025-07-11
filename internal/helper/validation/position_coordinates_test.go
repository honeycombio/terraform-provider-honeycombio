package validation_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"

	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/validation"
)

func Test_RequireBothCoordinatesValidator(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name        string
		value       types.Object
		expectError bool
	}

	attrTypes := map[string]attr.Type{
		"x_coordinate": types.Int64Type,
		"y_coordinate": types.Int64Type,
		"height":       types.Int64Type,
		"width":        types.Int64Type,
	}

	tests := []testCase{
		{
			name:  "unknown",
			value: types.ObjectUnknown(attrTypes),
		},
		{
			name:  "null",
			value: types.ObjectNull(attrTypes),
		},
		{
			name: "both coordinates set",
			value: types.ObjectValueMust(attrTypes, map[string]attr.Value{
				"x_coordinate": types.Int64Value(10),
				"y_coordinate": types.Int64Value(20),
				"height":       types.Int64Null(),
				"width":        types.Int64Null(),
			}),
		},
		{
			name: "neither coordinate set",
			value: types.ObjectValueMust(attrTypes, map[string]attr.Value{
				"x_coordinate": types.Int64Null(),
				"y_coordinate": types.Int64Null(),
				"height":       types.Int64Value(100),
				"width":        types.Int64Value(200),
			}),
		},
		{
			name: "only x coordinate set",
			value: types.ObjectValueMust(attrTypes, map[string]attr.Value{
				"x_coordinate": types.Int64Value(10),
				"y_coordinate": types.Int64Null(),
				"height":       types.Int64Null(),
				"width":        types.Int64Null(),
			}),
			expectError: true,
		},
		{
			name: "only y coordinate set",
			value: types.ObjectValueMust(attrTypes, map[string]attr.Value{
				"x_coordinate": types.Int64Null(),
				"y_coordinate": types.Int64Value(20),
				"height":       types.Int64Null(),
				"width":        types.Int64Null(),
			}),
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			request := validator.ObjectRequest{
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				ConfigValue:    tc.value,
			}
			response := validator.ObjectResponse{}

			validation.RequireBothCoordinates().ValidateObject(context.Background(), request, &response)

			assert.Equal(t,
				tc.expectError,
				response.Diagnostics.HasError(),
				"unexpected error: %s", response.Diagnostics,
			)
		})
	}
}
