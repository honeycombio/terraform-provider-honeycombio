package validation

import (
	"context"
	"fmt"
	"net/mail"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.String = isEmailAddressValidator{}

type isEmailAddressValidator struct{}

func (v isEmailAddressValidator) Description(_ context.Context) string {
	return "value must be a bare email address, without a name, angle brackets or spaces"
}

func (v isEmailAddressValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v isEmailAddressValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	// the API accepts an address only when parsing gives back exactly what was sent
	email := request.ConfigValue.ValueString()
	if addr, err := mail.ParseAddress(email); err != nil || addr.Address != email {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			request.Path,
			v.Description(ctx),
			fmt.Sprintf("%q", email),
		))
	}
}

// IsEmailAddress returns a validator which ensures that a configured value
// is an email address the Honeycomb API accepts, such as "alice@example.com".
//
// Null (unconfigured) and unknown (known after apply) values are skipped.
func IsEmailAddress() validator.String {
	return isEmailAddressValidator{}
}
