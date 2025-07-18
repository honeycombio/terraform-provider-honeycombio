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

func Test_RequireConsistentPanelPositionsValidator(t *testing.T) {
	t.Parallel()

	positionAttrTypes := map[string]attr.Type{
		"x_coordinate": types.Int64Type,
		"y_coordinate": types.Int64Type,
		"height":       types.Int64Type,
		"width":        types.Int64Type,
	}

	panelAttrTypes := map[string]attr.Type{
		"type": types.StringType,
		"position": types.ObjectType{
			AttrTypes: positionAttrTypes,
		},
		"query_panel": types.ListType{
			ElemType: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"query_id":            types.StringType,
					"query_annotation_id": types.StringType,
				},
			},
		},
		"slo_panel": types.ListType{
			ElemType: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"slo_id": types.StringType,
				},
			},
		},
	}

	panelListType := types.ListType{
		ElemType: types.ObjectType{
			AttrTypes: panelAttrTypes,
		},
	}

	tests := []struct {
		name        string
		value       types.List
		expectError bool
	}{
		{
			name:  "unknown",
			value: types.ListUnknown(panelListType.ElemType),
		},
		{
			name:  "null",
			value: types.ListNull(panelListType.ElemType),
		},
		{
			name:  "empty list",
			value: types.ListValueMust(panelListType.ElemType, []attr.Value{}),
		},
		{
			name: "single panel with position",
			value: types.ListValueMust(panelListType.ElemType, []attr.Value{
				types.ObjectValueMust(panelAttrTypes, map[string]attr.Value{
					"type": types.StringValue("query"),
					"position": types.ObjectValueMust(positionAttrTypes, map[string]attr.Value{
						"x_coordinate": types.Int64Value(10),
						"y_coordinate": types.Int64Value(20),
						"height":       types.Int64Null(),
						"width":        types.Int64Null(),
					}),
					"query_panel": types.ListNull(types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"query_id":            types.StringType,
							"query_annotation_id": types.StringType,
						},
					}),
					"slo_panel": types.ListNull(types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"slo_id": types.StringType,
						},
					}),
				}),
			}),
		},
		{
			name: "single panel without position",
			value: types.ListValueMust(panelListType.ElemType, []attr.Value{
				types.ObjectValueMust(panelAttrTypes, map[string]attr.Value{
					"type":     types.StringValue("query"),
					"position": types.ObjectNull(positionAttrTypes),
					"query_panel": types.ListNull(types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"query_id":            types.StringType,
							"query_annotation_id": types.StringType,
						},
					}),
					"slo_panel": types.ListNull(types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"slo_id": types.StringType,
						},
					}),
				}),
			}),
		},
		{
			name: "all panels have positions",
			value: types.ListValueMust(panelListType.ElemType, []attr.Value{
				types.ObjectValueMust(panelAttrTypes, map[string]attr.Value{
					"type": types.StringValue("query"),
					"position": types.ObjectValueMust(positionAttrTypes, map[string]attr.Value{
						"x_coordinate": types.Int64Value(10),
						"y_coordinate": types.Int64Value(20),
						"height":       types.Int64Null(),
						"width":        types.Int64Null(),
					}),
					"query_panel": types.ListNull(types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"query_id":            types.StringType,
							"query_annotation_id": types.StringType,
						},
					}),
					"slo_panel": types.ListNull(types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"slo_id": types.StringType,
						},
					}),
				}),
				types.ObjectValueMust(panelAttrTypes, map[string]attr.Value{
					"type": types.StringValue("slo"),
					"position": types.ObjectValueMust(positionAttrTypes, map[string]attr.Value{
						"x_coordinate": types.Int64Value(30),
						"y_coordinate": types.Int64Value(40),
						"height":       types.Int64Null(),
						"width":        types.Int64Null(),
					}),
					"query_panel": types.ListNull(types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"query_id":            types.StringType,
							"query_annotation_id": types.StringType,
						},
					}),
					"slo_panel": types.ListNull(types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"slo_id": types.StringType,
						},
					}),
				}),
			}),
		},
		{
			name: "no panels have positions",
			value: types.ListValueMust(panelListType.ElemType, []attr.Value{
				types.ObjectValueMust(panelAttrTypes, map[string]attr.Value{
					"type":     types.StringValue("query"),
					"position": types.ObjectNull(positionAttrTypes),
					"query_panel": types.ListNull(types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"query_id":            types.StringType,
							"query_annotation_id": types.StringType,
						},
					}),
					"slo_panel": types.ListNull(types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"slo_id": types.StringType,
						},
					}),
				}),
				types.ObjectValueMust(panelAttrTypes, map[string]attr.Value{
					"type":     types.StringValue("slo"),
					"position": types.ObjectNull(positionAttrTypes),
					"query_panel": types.ListNull(types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"query_id":            types.StringType,
							"query_annotation_id": types.StringType,
						},
					}),
					"slo_panel": types.ListNull(types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"slo_id": types.StringType,
						},
					}),
				}),
			}),
		},
		{
			name: "mixed panels - some with positions, some without",
			value: types.ListValueMust(panelListType.ElemType, []attr.Value{
				types.ObjectValueMust(panelAttrTypes, map[string]attr.Value{
					"type": types.StringValue("query"),
					"position": types.ObjectValueMust(positionAttrTypes, map[string]attr.Value{
						"x_coordinate": types.Int64Value(10),
						"y_coordinate": types.Int64Value(20),
						"height":       types.Int64Null(),
						"width":        types.Int64Null(),
					}),
					"query_panel": types.ListNull(types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"query_id":            types.StringType,
							"query_annotation_id": types.StringType,
						},
					}),
					"slo_panel": types.ListNull(types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"slo_id": types.StringType,
						},
					}),
				}),
				types.ObjectValueMust(panelAttrTypes, map[string]attr.Value{
					"type":     types.StringValue("slo"),
					"position": types.ObjectNull(positionAttrTypes),
					"query_panel": types.ListNull(types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"query_id":            types.StringType,
							"query_annotation_id": types.StringType,
						},
					}),
					"slo_panel": types.ListNull(types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"slo_id": types.StringType,
						},
					}),
				}),
			}),
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			request := validator.ListRequest{
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				ConfigValue:    tc.value,
			}
			response := validator.ListResponse{}

			validation.RequireConsistentPanelPositions().ValidateList(context.Background(), request, &response)

			assert.Equal(t,
				tc.expectError,
				response.Diagnostics.HasError(),
				"unexpected error: %s", response.Diagnostics,
			)
		})
	}
}
