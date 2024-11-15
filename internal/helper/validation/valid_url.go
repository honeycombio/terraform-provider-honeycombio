package validation

import (
	"context"
	"fmt"
	"net/url"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.String = isValidURLValidator{}

type isValidURLValidator struct {
	schemes []string
}

func (v isValidURLValidator) Description(_ context.Context) string {
	return "value must be a valid URL"
}

func (v isValidURLValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v isValidURLValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	u, err := url.Parse(request.ConfigValue.ValueString())
	if err != nil {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			request.Path,
			v.Description(ctx),
			fmt.Sprintf("%q: %s", request.ConfigValue.ValueString(), err.Error()),
		))
		return
	}

	if u.Host == "" {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			request.Path,
			v.Description(ctx)+" but is missing a host",
			request.ConfigValue.ValueString(),
		))
	}

	if !slices.Contains(v.schemes, u.Scheme) {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			request.Path,
			fmt.Sprintf("%s with a schema of %q", v.Description(ctx), strings.Join(v.schemes, ",")),
			request.ConfigValue.ValueString(),
		))
	}
}

// IsURLWithHTTPorHTTPS returns an AttributeValidator which ensures that any configured
// attribute value is a valid HTTP or HTTPS URL.
//
// Null (unconfigured) and unknown (known after apply) values are skipped.
func IsURLWithHTTPorHTTPS() validator.String {
	return isValidURLValidator{[]string{"http", "https"}}
}
