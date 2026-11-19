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

func Test_IsEmailAddress(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tests := map[string]struct {
		val         types.String
		expectError bool
	}{
		"unknown":             {val: types.StringUnknown()},
		"null":                {val: types.StringNull()},
		"plain":               {val: types.StringValue("alice@example.com")},
		"upper case":          {val: types.StringValue("Alice@Example.COM")},
		"plus address":        {val: types.StringValue("alice+ops@example.com")},
		"empty":               {val: types.StringValue(""), expectError: true},
		"no at sign":          {val: types.StringValue("alice.example.com"), expectError: true},
		"with a name":         {val: types.StringValue("Alice <alice@example.com>"), expectError: true},
		"in angle brackets":   {val: types.StringValue("<alice@example.com>"), expectError: true},
		"leading space":       {val: types.StringValue(" alice@example.com"), expectError: true},
		"trailing space":      {val: types.StringValue("alice@example.com "), expectError: true},
		"space in local part": {val: types.StringValue("alice smith@example.com"), expectError: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			req := validator.StringRequest{Path: path.Root("email"), ConfigValue: tc.val}
			resp := validator.StringResponse{}
			validation.IsEmailAddress().ValidateString(ctx, req, &resp)

			assert.Equal(t, tc.expectError, resp.Diagnostics.HasError(), resp.Diagnostics)
		})
	}
}
