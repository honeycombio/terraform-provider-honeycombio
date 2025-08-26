package validation_test

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// PanelTestAttributeTypes provides common attribute type definitions for panel validation tests
type PanelTestAttributeTypes struct {
	Position   map[string]attr.Type
	QueryPanel map[string]attr.Type
	SLOPanel   map[string]attr.Type
	TextPanel  map[string]attr.Type
	Panel      map[string]attr.Type
	PanelList  types.ListType
}

// NewPanelTestAttributeTypes creates and returns the standard attribute types used across panel validation tests
func NewPanelTestAttributeTypes() *PanelTestAttributeTypes {
	positionAttrTypes := map[string]attr.Type{
		"x_coordinate": types.Int64Type,
		"y_coordinate": types.Int64Type,
		"height":       types.Int64Type,
		"width":        types.Int64Type,
	}

	queryPanelAttrTypes := map[string]attr.Type{
		"query_id":            types.StringType,
		"query_annotation_id": types.StringType,
	}

	sloPanelAttrTypes := map[string]attr.Type{
		"slo_id": types.StringType,
	}

	textPanelAttrTypes := map[string]attr.Type{
		"content": types.StringType,
	}

	panelAttrTypes := map[string]attr.Type{
		"type": types.StringType,
		"position": types.ObjectType{
			AttrTypes: positionAttrTypes,
		},
		"query_panel": types.ListType{
			ElemType: types.ObjectType{
				AttrTypes: queryPanelAttrTypes,
			},
		},
		"slo_panel": types.ListType{
			ElemType: types.ObjectType{
				AttrTypes: sloPanelAttrTypes,
			},
		},
		"text_panel": types.ListType{
			ElemType: types.ObjectType{
				AttrTypes: textPanelAttrTypes,
			},
		},
	}

	panelListType := types.ListType{
		ElemType: types.ObjectType{
			AttrTypes: panelAttrTypes,
		},
	}

	return &PanelTestAttributeTypes{
		Position:   positionAttrTypes,
		QueryPanel: queryPanelAttrTypes,
		SLOPanel:   sloPanelAttrTypes,
		TextPanel:  textPanelAttrTypes,
		Panel:      panelAttrTypes,
		PanelList:  panelListType,
	}
}
