package validation

import (
	"context"
	"fmt"
	"math"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.Int64 = divisibleByValidator{}

type divisibleByValidator struct {
	divisor int64
}

func (validator divisibleByValidator) Description(_ context.Context) string {
	return fmt.Sprintf("value must be divisible by %d", validator.divisor)
}

func (validator divisibleByValidator) MarkdownDescription(ctx context.Context) string {
	return validator.Description(ctx)
}

func (validator divisibleByValidator) ValidateInt64(ctx context.Context, request validator.Int64Request, response *validator.Int64Response) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	if math.Mod(float64(request.ConfigValue.ValueInt64()), float64(validator.divisor)) != 0 {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			request.Path,
			validator.Description(ctx),
			fmt.Sprintf("%d", request.ConfigValue.ValueInt64()),
		))
	}
}

// Int64DivisbileBy returns an AttributeValidator which ensures that any configured
// attribute value:
//
//   - Is a number, which can be represented by a 64-bit integer.
//   - Is evenly divisible by a given number.
//
// Null (unconfigured) and unknown (known after apply) values are skipped.
func Int64DivisibleBy(v int64) validator.Int64 {
	return divisibleByValidator{divisor: v}
}
