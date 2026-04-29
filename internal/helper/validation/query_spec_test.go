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

func Test_QuerySpecValidator(t *testing.T) {
	t.Parallel()

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
		"valid COUNT without column": {
			val: types.StringValue(`{"calculations": [{"op": "COUNT"}]}`),
		},
		"valid COUNT with column": {
			val: types.StringValue(`{"calculations": [{"op": "COUNT", "column": "app.cumulative"}]}`),
		},
		"valid COUNT_DATAPOINTS without column": {
			val: types.StringValue(`{"calculations": [{"op": "COUNT_DATAPOINTS"}]}`),
		},
		"valid COUNT_DATAPOINTS with column": {
			val: types.StringValue(`{"calculations": [{"op": "COUNT_DATAPOINTS", "column": "app.cumulative"}]}`),
		},
		"valid HISTOGRAM_COUNT": {
			val: types.StringValue(`{"calculations": [{"op": "HISTOGRAM_COUNT", "column": "request.duration"}]}`),
		},
		"invalid json": {
			val:         types.StringValue("whoop"),
			expectError: true,
		},
		"invalid query spec": {
			val:         types.StringValue(`{"foo": "bar"}`),
			expectError: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			request := validator.StringRequest{
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				ConfigValue:    test.val,
			}
			response := validator.StringResponse{}
			validation.ValidQuerySpec().ValidateString(context.Background(), request, &response)

			assert.Equal(t,
				test.expectError,
				response.Diagnostics.HasError(),
				"unexpected error: %s", response.Diagnostics,
			)
		})
	}
}
