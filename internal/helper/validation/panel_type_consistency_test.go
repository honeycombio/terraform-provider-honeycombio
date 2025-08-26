package validation_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"

	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/validation"
)

func Test_RequirePanelTypeConsistencyValidator(t *testing.T) {
	t.Parallel()

	// Use shared test attribute types
	testTypes := NewPanelTestAttributeTypes()
	positionAttrTypes := testTypes.Position
	queryPanelAttrTypes := testTypes.QueryPanel
	sloPanelAttrTypes := testTypes.SLOPanel
	textPanelAttrTypes := testTypes.TextPanel
	panelAttrTypes := testTypes.Panel
	panelListType := testTypes.PanelList

	tests := []struct {
		name        string
		value       types.List
		expectError bool
		description string
	}{
		{
			name:        "null value",
			value:       types.ListNull(panelListType.ElemType),
			expectError: false,
			description: "null values should be skipped",
		},
		{
			name:        "unknown value",
			value:       types.ListUnknown(panelListType.ElemType),
			expectError: false,
			description: "unknown values should be skipped",
		},
		{
			name:        "empty list",
			value:       types.ListValueMust(panelListType.ElemType, []attr.Value{}),
			expectError: false,
			description: "empty lists should be valid",
		},
		{
			name: "valid query panel",
			value: types.ListValueMust(panelListType.ElemType, []attr.Value{
				types.ObjectValueMust(panelAttrTypes, map[string]attr.Value{
					"type":     types.StringValue("query"),
					"position": types.ObjectNull(positionAttrTypes),
					"query_panel": types.ListValueMust(
						types.ObjectType{AttrTypes: queryPanelAttrTypes},
						[]attr.Value{
							types.ObjectValueMust(queryPanelAttrTypes, map[string]attr.Value{
								"query_id":            types.StringValue("query123"),
								"query_annotation_id": types.StringValue("annotation123"),
							}),
						},
					),
					"slo_panel":  types.ListNull(types.ObjectType{AttrTypes: sloPanelAttrTypes}),
					"text_panel": types.ListNull(types.ObjectType{AttrTypes: textPanelAttrTypes}),
				}),
			}),
			expectError: false,
			description: "valid query panel should pass validation",
		},
		{
			name: "query panel without query_panel config",
			value: types.ListValueMust(panelListType.ElemType, []attr.Value{
				types.ObjectValueMust(panelAttrTypes, map[string]attr.Value{
					"type":        types.StringValue("query"),
					"position":    types.ObjectNull(positionAttrTypes),
					"query_panel": types.ListNull(types.ObjectType{AttrTypes: queryPanelAttrTypes}),
					"slo_panel":   types.ListNull(types.ObjectType{AttrTypes: sloPanelAttrTypes}),
					"text_panel":  types.ListNull(types.ObjectType{AttrTypes: textPanelAttrTypes}),
				}),
			}),
			expectError: true,
			description: "query panel without query_panel config should fail",
		},
		{
			name: "query panel with slo_panel config",
			value: types.ListValueMust(panelListType.ElemType, []attr.Value{
				types.ObjectValueMust(panelAttrTypes, map[string]attr.Value{
					"type":     types.StringValue("query"),
					"position": types.ObjectNull(positionAttrTypes),
					"query_panel": types.ListValueMust(
						types.ObjectType{AttrTypes: queryPanelAttrTypes},
						[]attr.Value{
							types.ObjectValueMust(queryPanelAttrTypes, map[string]attr.Value{
								"query_id":            types.StringValue("query123"),
								"query_annotation_id": types.StringValue("annotation123"),
							}),
						},
					),
					"slo_panel": types.ListValueMust(
						types.ObjectType{AttrTypes: sloPanelAttrTypes},
						[]attr.Value{
							types.ObjectValueMust(sloPanelAttrTypes, map[string]attr.Value{
								"slo_id": types.StringValue("slo123"),
							}),
						},
					),
					"text_panel": types.ListNull(types.ObjectType{AttrTypes: textPanelAttrTypes}),
				}),
			}),
			expectError: true,
			description: "query panel with slo_panel config should fail",
		},
		{
			name: "query panel with empty query_panel config",
			value: types.ListValueMust(panelListType.ElemType, []attr.Value{
				types.ObjectValueMust(panelAttrTypes, map[string]attr.Value{
					"type":     types.StringValue("query"),
					"position": types.ObjectNull(positionAttrTypes),
					"query_panel": types.ListValueMust(
						types.ObjectType{AttrTypes: queryPanelAttrTypes},
						[]attr.Value{},
					),
					"slo_panel":  types.ListNull(types.ObjectType{AttrTypes: sloPanelAttrTypes}),
					"text_panel": types.ListNull(types.ObjectType{AttrTypes: textPanelAttrTypes}),
				}),
			}),
			expectError: true,
			description: "query panel with empty query_panel config should fail",
		},
		{
			name: "invalid panel type",
			value: types.ListValueMust(panelListType.ElemType, []attr.Value{
				types.ObjectValueMust(panelAttrTypes, map[string]attr.Value{
					"type":        types.StringValue("invalid"),
					"position":    types.ObjectNull(positionAttrTypes),
					"query_panel": types.ListNull(types.ObjectType{AttrTypes: queryPanelAttrTypes}),
					"slo_panel":   types.ListNull(types.ObjectType{AttrTypes: sloPanelAttrTypes}),
					"text_panel":  types.ListNull(types.ObjectType{AttrTypes: textPanelAttrTypes}),
				}),
			}),
			expectError: true,
			description: "invalid panel type should fail",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := validator.ListRequest{
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				ConfigValue:    test.value,
			}
			response := &validator.ListResponse{}

			validation.RequirePanelTypeConsistency().ValidateList(context.Background(), request, response)

			if test.expectError {
				assert.True(t, response.Diagnostics.HasError(), "Expected error for test case: %s", test.description)
			} else {
				assert.False(t, response.Diagnostics.HasError(), "Expected no error for test case: %s. Got: %v", test.description, response.Diagnostics.Errors())
			}
		})
	}
}
