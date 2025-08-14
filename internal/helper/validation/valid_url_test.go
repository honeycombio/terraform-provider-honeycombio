package validation_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"

	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/validation"
)

func Test_IsURLWithHTTPorHTTPS(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	type testCase struct {
		val         types.String
		expectError bool
	}
	tests := map[string]testCase{
		"unknown": {
			val: types.StringUnknown(),
		},
		"null": {
			val: types.StringNull(),
		},
		"valid http": {
			val: types.StringValue("http://sub.example.com/a/b/c/d?e=f#g"),
		},
		"valid https": {
			val: types.StringValue("https://sub.example.com/a/b/c/d?e=f#g"),
		},
		"empty": {
			val:         types.StringValue(""),
			expectError: true,
		},
		"garbage": {
			val:         types.StringValue("not-a-url"),
			expectError: true,
		},
		"missing host": {
			val:         types.StringValue("http:///a/b/c/d?e=f#g"),
			expectError: true,
		},
		"invalid scheme": {
			val:         types.StringValue("ftp://sub.example.com/"),
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
			validation.IsURLWithHTTPorHTTPS().ValidateString(ctx, request, &response)

			assert.Equal(t,
				test.expectError,
				response.Diagnostics.HasError(),
				"unexpected error: %s", response.Diagnostics,
			)
		})
	}
}
