package features

import "github.com/hashicorp/terraform-plugin-framework/types"

// Features represents provider-level features.
type Features struct {
	Column  FeaturesColumn
	Dataset FeaturesDataset
}

// FeaturesColumn represents column-specific features.
type FeaturesColumn struct {
	// ImportOnConflict controls whether to import an existing column if a create
	// operation fails due to a conflict (HTTP 409).
	ImportOnConflict bool
}

// FeaturesColumnModel represents column-specific features for Terraform schema.
type FeaturesColumnModel struct {
	ImportOnConflict types.Bool `tfsdk:"import_on_conflict"`
}

// FeaturesDataset represents dataset-specific features.
type FeaturesDataset struct {
	// ImportOnConflict controls whether to import an existing dataset if a create
	// operation results in an HTTP 200 instead of an HTTP 201.
	ImportOnConflict bool
}

type FeaturesDatasetModel struct {
	ImportOnConflict types.Bool `tfsdk:"import_on_conflict"`
}

type Model struct {
	Column  []FeaturesColumnModel  `tfsdk:"column"`
	Dataset []FeaturesDatasetModel `tfsdk:"dataset"`
}

// Parse converts a Terraform model to internal Features representation for
// plugin-based providers while handling default values.
func Parse(m []Model) *Features {
	result := DefaultFeatures()
	if len(m) == 0 {
		return result
	}
	features := m[0]

	// parse column features
	if len(features.Column) > 0 {
		columnFeatures := features.Column[0]
		if !columnFeatures.ImportOnConflict.IsNull() && !columnFeatures.ImportOnConflict.IsUnknown() {
			result.Column.ImportOnConflict = columnFeatures.ImportOnConflict.ValueBool()
		}
	}

	// parse dataset features
	if len(features.Dataset) > 0 {
		datasetFeatures := features.Dataset[0]
		if !datasetFeatures.ImportOnConflict.IsNull() && !datasetFeatures.ImportOnConflict.IsUnknown() {
			result.Dataset.ImportOnConflict = datasetFeatures.ImportOnConflict.ValueBool()
		}
	}

	return result
}
