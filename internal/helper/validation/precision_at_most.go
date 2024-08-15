package validation

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
)

var _ validator.Float64 = precisionAtMostValidator{}

// precisionAtMostValidator validates that a float attribute's precision is at most a certain value.
type precisionAtMostValidator struct {
	max int64
}

func (validator precisionAtMostValidator) Description(_ context.Context) string {
	return fmt.Sprintf("precision for value must be at most %d", validator.max)
}

func (validator precisionAtMostValidator) MarkdownDescription(ctx context.Context) string {
	return validator.Description(ctx)
}

// ValidateFloat64 performs the validation.
func (validator precisionAtMostValidator) ValidateFloat64(ctx context.Context, request validator.Float64Request, response *validator.Float64Response) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	value := request.ConfigValue.ValueFloat64()
	valueAsString := helper.FloatToPercentString(value)
	splitValue := strings.Split(valueAsString, ".")

	// If length is less than 2, then we don't have a decimal point
	// so we know the precision is valid
	if len(splitValue) < 2 {
		return
	}

	afterDecimal := splitValue[1]
	valuePrecision := int64(len(afterDecimal))
	if valuePrecision > validator.max {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			request.Path,
			validator.Description(ctx),
			fmt.Sprintf("%f", value),
		))
	}
}

// Float64PrecisionAtMost returns an AttributeValidator which ensures that any configured
// attribute value has a precision which is at most the given max
func Float64PrecisionAtMost(m int64) validator.Float64 {
	return precisionAtMostValidator{
		max: m,
	}
}
