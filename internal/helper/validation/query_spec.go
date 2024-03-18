package validation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
)

var _ validator.String = querySpecValidator{}

type querySpecValidator struct{}

func (v querySpecValidator) Description(_ context.Context) string {
	return "value must be a valid Honeycomb Query Specification"
}

func (v querySpecValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v querySpecValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	var q client.QuerySpec
	dec := json.NewDecoder(bytes.NewReader([]byte(request.ConfigValue.ValueString())))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&q); err != nil {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			request.Path,
			v.Description(ctx),
			fmt.Sprintf("%q: %s", request.ConfigValue.ValueString(), err.Error()),
		))
	}
}

// ValidQuerySpec determines if the provided JSON is a valid Honeycomb Query Specification
func ValidQuerySpec() validator.String {
	return querySpecValidator{}
}
