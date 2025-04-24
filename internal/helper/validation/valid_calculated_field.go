package validation

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	dcparser "github.com/honeycombio/honeycomb-derived-column-validator/pkg/parser"
)

var _ validator.String = isValidCalculatedFieldValidator{}

type isValidCalculatedFieldValidator struct{}

func (v isValidCalculatedFieldValidator) Description(_ context.Context) string {
	return "expression must be a valid calculated field"
}

func (v isValidCalculatedFieldValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v isValidCalculatedFieldValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	if _, err := dcparser.ANTLRParse(request.ConfigValue.ValueString(), false); err != nil {
		response.Diagnostics.Append(
			validatordiag.InvalidAttributeValueDiagnostic(
				request.Path,
				v.Description(ctx),
				fmt.Sprintf("%q: %s", request.ConfigValue.ValueString(), err),
			),
		)
	}
}

// IsValidCalculatedField returns an AttributeValidator which ensures that any configured
// attribute value is a valid calculated field expression.
//
// Null (unconfigured) and unknown (known after apply) values are skipped.
func IsValidCalculatedField() validator.String {
	return isValidCalculatedFieldValidator{}
}
