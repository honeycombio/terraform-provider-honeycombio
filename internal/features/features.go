package features

import "github.com/hashicorp/terraform-plugin-framework/types"

// Features represents provider-level features.
type Features struct {
	Column FeaturesColumn
}

// FeaturesColumn represents column-specific features.
type FeaturesColumn struct {
	ImportOnConflict bool
}

// FeaturesColumnModel represents column-specific features for Terraform schema.
type FeaturesColumnModel struct {
	ImportOnConflict types.Bool `tfsdk:"import_on_conflict"`
}

type Model struct {
	Column []FeaturesColumnModel `tfsdk:"column"`
}

// Parse converts a Terraform model to internal features representation for
// plugin-based providers. Use ParseSDKResourceData for SDK-based providers.
// It starts with default values and applies any explicitly configured overrides.
func Parse(m Model) *Features {
	features := DefaultFeatures()

	if len(m.Column) > 0 {
		columnFeatures := m.Column[0]
		if !columnFeatures.ImportOnConflict.IsNull() && !columnFeatures.ImportOnConflict.IsUnknown() {
			features.Column.ImportOnConflict = columnFeatures.ImportOnConflict.ValueBool()
		}
	}

	return &features
}

// ParseSDKResourceData parses features from SDK v2 ResourceData.
// Use the Model.Parse method for plugin-based providers.
// It starts with default values and applies any explicitly configured overrides.
func ParseSDKResourceData(d interface{}) *Features {
	features := DefaultFeatures()

	type ResourceData interface {
		GetOk(key string) (interface{}, bool)
	}

	rd, ok := d.(ResourceData)
	if !ok {
		return &features
	}

	if featuresRaw, ok := rd.GetOk("features"); ok {
		featuresList := featuresRaw.([]interface{})

		if len(featuresList) > 0 {
			featuresMap := featuresList[0].(map[string]interface{})
			if columnRaw, ok := featuresMap["column"]; ok {
				columnList := columnRaw.([]interface{})
				if len(columnList) > 0 {
					columnMap := columnList[0].(map[string]interface{})
					if importOnConflict, ok := columnMap["import_on_conflict"]; ok {
						features.Column.ImportOnConflict = importOnConflict.(bool)
					}
				}
			}
		}
	}

	return &features
}
