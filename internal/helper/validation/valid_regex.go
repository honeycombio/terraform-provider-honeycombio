package validation

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.String = isValidRegExpValidator{}

type isValidRegExpValidator struct{}

func (v isValidRegExpValidator) Description(_ context.Context) string {
	return "value must be a valid regular expression"
}

func (v isValidRegExpValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v isValidRegExpValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	if _, err := regexp.Compile(request.ConfigValue.ValueString()); err != nil {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			request.Path,
			v.Description(ctx),
			fmt.Sprintf("%q: %s", request.ConfigValue.ValueString(), err.Error()),
		))
	}
}

// IsValidRegExp returns an AttributeValidator which ensures that any configured
// attribute value is a valid regular expression.
//
// Null (unconfigured) and unknown (known after apply) values are skipped.
func IsValidRegExp() validator.String {
	return isValidRegExpValidator{}
}
