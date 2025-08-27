package validation

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ validator.List = panelTypeConsistencyValidator{}

// panelTypeConsistencyValidator validates that when a panel "type" is provided,
// the corresponding panel_type is also provided and no other panel_types are present.
type panelTypeConsistencyValidator struct{}

func (v panelTypeConsistencyValidator) Description(_ context.Context) string {
	return "when a panel type is specified, only the corresponding panel configuration should be provided"
}

func (v panelTypeConsistencyValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v panelTypeConsistencyValidator) ValidateList(ctx context.Context, request validator.ListRequest, response *validator.ListResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	elements := request.ConfigValue.Elements()
	if len(elements) == 0 {
		return
	}

	for i, element := range elements {
		if element.IsNull() || element.IsUnknown() {
			continue
		}

		panelObj, ok := element.(types.Object)
		if !ok {
			continue
		}

		attrs := panelObj.Attributes()

		// Get the panel type
		typeAttr, exists := attrs["type"]
		if !exists || typeAttr.IsNull() || typeAttr.IsUnknown() {
			continue
		}

		typeStr, ok := typeAttr.(types.String)
		if !ok {
			continue
		}

		panelType := typeStr.ValueString()

		// Define the mapping between panel types and their corresponding panel configuration blocks
		panelTypeMapping := map[string]string{
			"query": "query_panel",
			"slo":   "slo_panel",
			"text":  "text_panel",
		}

		expectedPanelAttr, validType := panelTypeMapping[panelType]
		if !validType {
			response.Diagnostics.Append(
				validatordiag.InvalidAttributeValueDiagnostic(
					request.Path.AtListIndex(i).AtName("type"),
					v.Description(ctx),
					fmt.Sprintf("Invalid panel type '%s'. Valid types are: query, slo, text.", panelType),
				),
			)
			continue
		}

		// Check that the expected panel configuration is provided
		expectedPanelConfig, exists := attrs[expectedPanelAttr]
		if !exists || expectedPanelConfig.IsNull() || expectedPanelConfig.IsUnknown() {
			response.Diagnostics.Append(
				validatordiag.InvalidAttributeValueDiagnostic(
					request.Path.AtListIndex(i),
					v.Description(ctx),
					fmt.Sprintf("Panel type '%s' requires '%s' configuration to be provided.", panelType, expectedPanelAttr),
				),
			)
			continue
		}

		// Check that the expected panel configuration is not empty
		expectedPanelList, ok := expectedPanelConfig.(types.List)
		if ok && !expectedPanelList.IsNull() && !expectedPanelList.IsUnknown() {
			if len(expectedPanelList.Elements()) == 0 {
				response.Diagnostics.Append(
					validatordiag.InvalidAttributeValueDiagnostic(
						request.Path.AtListIndex(i).AtName(expectedPanelAttr),
						v.Description(ctx),
						fmt.Sprintf("Panel type '%s' requires '%s' configuration to be non-empty.", panelType, expectedPanelAttr),
					),
				)
				continue
			}
		}

		// Check that no other panel configurations are provided
		for otherType, otherPanelAttr := range panelTypeMapping {
			if otherType == panelType {
				continue // Skip the expected panel type
			}

			otherPanelConfig, exists := attrs[otherPanelAttr]
			if !exists {
				continue
			}

			// Check if the other panel configuration is set and not empty
			if !otherPanelConfig.IsNull() && !otherPanelConfig.IsUnknown() {
				otherPanelList, ok := otherPanelConfig.(types.List)
				if ok && len(otherPanelList.Elements()) > 0 {
					response.Diagnostics.Append(
						validatordiag.InvalidAttributeValueDiagnostic(
							request.Path.AtListIndex(i).AtName(otherPanelAttr),
							v.Description(ctx),
							fmt.Sprintf("Panel type '%s' should not have '%s' configuration. Only '%s' should be provided.", panelType, otherPanelAttr, expectedPanelAttr),
						),
					)
				}
			}
		}
	}
}

// RequirePanelTypeConsistency returns a ListValidator which ensures that when a panel "type" is provided,
// the corresponding panel_type is also provided and no other panel_types are present.
//
// Null (unconfigured) and unknown (known after apply) values are skipped.
func RequirePanelTypeConsistency() validator.List {
	return panelTypeConsistencyValidator{}
}
