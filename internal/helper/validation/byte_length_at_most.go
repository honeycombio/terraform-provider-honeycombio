package validation

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.String = byteLengthAtMostValidator{}

// byteLengthAtMostValidator validates that a string attribute is at most
// a certain number of bytes long.
//
// This differs from stringvalidator.LengthAtMost and UTF8LengthAtMost
// which both count characters, so multi-byte characters pass those checks
// while still exceeding a limit the API enforces in bytes.
type byteLengthAtMostValidator struct {
	max int
}

func (v byteLengthAtMostValidator) Description(_ context.Context) string {
	return fmt.Sprintf("size must be at most %d bytes", v.max)
}

func (v byteLengthAtMostValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v byteLengthAtMostValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	if n := len(req.ConfigValue.ValueString()); n > v.max {
		resp.Diagnostics.Append(validatordiag.InvalidAttributeValueLengthDiagnostic(
			req.Path,
			v.Description(ctx),
			fmt.Sprintf("%d bytes", n),
		))
	}
}

// StringByteLengthAtMost returns a validator which ensures that any configured
// string value is at most the given number of bytes long.
func StringByteLengthAtMost(limit int) validator.String {
	return byteLengthAtMostValidator{max: limit}
}
