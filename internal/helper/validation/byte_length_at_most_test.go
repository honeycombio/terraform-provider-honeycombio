package validation_test

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"

	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/validation"
)

func Test_StringByteLengthAtMost(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		val         types.String
		limit       int
		expectError bool
	}{
		"unknown":                 {val: types.StringUnknown(), limit: 3},
		"null":                    {val: types.StringNull(), limit: 3},
		"empty":                   {val: types.StringValue(""), limit: 3},
		"at the limit":            {val: types.StringValue("abc"), limit: 3},
		"over the limit":          {val: types.StringValue("abcd"), limit: 3, expectError: true},
		"multi-byte at the limit": {val: types.StringValue("é"), limit: 2},
		// 2 characters, 4 bytes: a character count would let this through
		"multi-byte over the limit": {val: types.StringValue("éé"), limit: 3, expectError: true},
		"large at the limit":        {val: types.StringValue(strings.Repeat("a", 1<<20)), limit: 1 << 20},
		"large over the limit":      {val: types.StringValue(strings.Repeat("a", 1<<20+1)), limit: 1 << 20, expectError: true},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			req := validator.StringRequest{
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				ConfigValue:    test.val,
			}
			resp := &validator.StringResponse{}
			validation.StringByteLengthAtMost(test.limit).ValidateString(context.Background(), req, resp)

			assert.Equal(t, test.expectError, resp.Diagnostics.HasError(), resp.Diagnostics)
		})
	}
}
