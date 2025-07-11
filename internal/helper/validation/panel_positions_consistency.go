package validation

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ validator.List = panelPositionsConsistencyValidator{}

// panelPositionsConsistencyValidator validates that either all panels have positions or none of them do.
type panelPositionsConsistencyValidator struct{}

func (v panelPositionsConsistencyValidator) Description(_ context.Context) string {
	return "either all panels must have positions or none of them should have positions"
}

func (v panelPositionsConsistencyValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v panelPositionsConsistencyValidator) ValidateList(ctx context.Context, request validator.ListRequest, response *validator.ListResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	elements := request.ConfigValue.Elements()
	if len(elements) == 0 {
		return
	}

	var hasPositionCount int
	var noPositionCount int

	for _, element := range elements {
		if element.IsNull() || element.IsUnknown() {
			continue
		}

		panelObj, ok := element.(types.Object)
		if !ok {
			continue
		}

		attrs := panelObj.Attributes()
		position, exists := attrs["position"]
		if !exists {
			continue
		}

		// Check if position is set (not null and not unknown)
		if position.IsNull() || position.IsUnknown() {
			noPositionCount++
		} else {
			// Check if position object has any actual values set
			positionObj, ok := position.(types.Object)
			if !ok {
				noPositionCount++
				continue
			}

			positionAttrs := positionObj.Attributes()
			hasAnyCoordinate := false

			// Check if any coordinate is set
			for _, coordName := range []string{"x_coordinate", "y_coordinate"} {
				if coord, exists := positionAttrs[coordName]; exists {
					if !coord.IsNull() && !coord.IsUnknown() {
						hasAnyCoordinate = true
						break
					}
				}
			}

			if hasAnyCoordinate {
				hasPositionCount++
			} else {
				noPositionCount++
			}
		}
	}

	// If we have both panels with positions and panels without positions, that's invalid
	if hasPositionCount > 0 && noPositionCount > 0 {
		response.Diagnostics.Append(
			validatordiag.InvalidAttributeValueDiagnostic(
				request.Path,
				v.Description(ctx),
				"Found panels with positions and panels without positions. All panels must have consistent position configuration.",
			),
		)
	}
}

// RequireConsistentPanelPositions returns a ListValidator which ensures that either all panels
// have positions or none of them do.
//
// Null (unconfigured) and unknown (known after apply) values are skipped.
func RequireConsistentPanelPositions() validator.List {
	return panelPositionsConsistencyValidator{}
}
