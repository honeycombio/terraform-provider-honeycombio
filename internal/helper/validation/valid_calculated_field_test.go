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

func Test_IsValidCalculatedFieldValidator(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		val         types.String
		expectError bool
	}{
		"unknown": {
			val: types.StringUnknown(),
		},
		"null": {
			val: types.StringNull(),
		},
		"valid": {
			val: types.StringValue(`IF(AND(NOT(EXISTS($trace.parent_id)),EXISTS($duration_ms)),LTE($duration_ms,300))`),
		},
		"valid with infix": {
			val: types.StringValue(`IF(!EXISTS($trace.parent_id)) AND EXISTS($duration_ms), $duration_ms <= 300)`),
		},
		"mismatched input": {
			val:         types.StringValue(`BOOL(1`),
			expectError: true,
		},
		"invalid function": {
			val:         types.StringValue(`FOOBAR(1)`),
			expectError: true,
		},
		"extraneous input": {
			val:         types.StringValue(`IF(AND(NOT(EXISTS($trace.parent_id)),EXISTS($duration_ms)),LTE($duration_ms,300)),`),
			expectError: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			request := validator.StringRequest{
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				ConfigValue:    tc.val,
			}
			response := validator.StringResponse{}

			validation.IsValidCalculatedField().
				ValidateString(context.Background(), request, &response)

			assert.Equal(t,
				tc.expectError,
				response.Diagnostics.HasError(),
				"unexpected error: %s", response.Diagnostics,
			)
		})
	}
}
