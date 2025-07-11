package validation

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.Object = positionCoordinatesValidator{}

// positionCoordinatesValidator validates that both x and y coordinates are set when one is provided.
type positionCoordinatesValidator struct{}

func (v positionCoordinatesValidator) Description(_ context.Context) string {
	return "both x_coordinate and y_coordinate must be set when one is provided"
}

func (v positionCoordinatesValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v positionCoordinatesValidator) ValidateObject(ctx context.Context, request validator.ObjectRequest, response *validator.ObjectResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	attrs := request.ConfigValue.Attributes()

	x, xExists := attrs["x_coordinate"]
	y, yExists := attrs["y_coordinate"]

	if !xExists || !yExists {
		return
	}

	// Check if one coordinate is set but the other is not
	xIsSet := !x.IsNull() && !x.IsUnknown()
	yIsSet := !y.IsNull() && !y.IsUnknown()

	if xIsSet && !yIsSet {
		response.Diagnostics.Append(
			validatordiag.InvalidAttributeValueDiagnostic(
				request.Path.AtName("y_coordinate"),
				v.Description(ctx),
				"y_coordinate is required when x_coordinate is set",
			),
		)
	}

	if yIsSet && !xIsSet {
		response.Diagnostics.Append(
			validatordiag.InvalidAttributeValueDiagnostic(
				request.Path.AtName("x_coordinate"),
				v.Description(ctx),
				"x_coordinate is required when y_coordinate is set",
			),
		)
	}
}

// RequireBothCoordinates returns an ObjectValidator which ensures that both x_coordinate and y_coordinate
// are set when one is provided.
//
// Null (unconfigured) and unknown (known after apply) values are skipped.
func RequireBothCoordinates() validator.Object {
	return positionCoordinatesValidator{}
}
