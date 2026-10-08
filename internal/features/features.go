package features

import "github.com/hashicorp/terraform-plugin-framework/types"

// Features represents provider-level features.
type Features struct {
	Client       FeaturesClient
	Column       FeaturesColumn
	Dataset      FeaturesDataset
	Intelligence FeaturesIntelligence
}

// FeaturesClient represents API client-specific features.
type FeaturesClient struct {
	// ReadCaching controls whether reads of supported resource types are
	// served from a short-lived cache of the containing collection, so
	// large fleets don't issue one API request per resource read.
	// Currently supported: derived columns.
	ReadCaching bool
}

// FeaturesClientModel represents API client-specific features for Terraform schema.
type FeaturesClientModel struct {
	ReadCaching types.Bool `tfsdk:"read_caching"`
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

// FeaturesIntelligence represents Honeycomb Intelligence-specific features.
type FeaturesIntelligence struct {
	// Enabled indicates the team has Honeycomb Intelligence enabled,
	// unlocking features such as auto-investigation on burn alerts and triggers.
	Enabled bool
}

// FeaturesIntelligenceModel represents Intelligence features for Terraform schema.
type FeaturesIntelligenceModel struct {
	Enabled types.Bool `tfsdk:"enabled"`
}

type Model struct {
	Client       []FeaturesClientModel       `tfsdk:"client"`
	Column       []FeaturesColumnModel       `tfsdk:"column"`
	Dataset      []FeaturesDatasetModel      `tfsdk:"dataset"`
	Intelligence []FeaturesIntelligenceModel `tfsdk:"intelligence"`
}

// Parse converts a Terraform model to internal Features representation for
// plugin-based providers while handling default values.
func Parse(m []Model) *Features {
	result := DefaultFeatures()
	if len(m) == 0 {
		return result
	}
	features := m[0]

	// parse client features
	if len(features.Client) > 0 {
		clientFeatures := features.Client[0]
		if !clientFeatures.ReadCaching.IsNull() && !clientFeatures.ReadCaching.IsUnknown() {
			result.Client.ReadCaching = clientFeatures.ReadCaching.ValueBool()
		}
	}

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

	// parse intelligence features
	if len(features.Intelligence) > 0 {
		intelligenceFeatures := features.Intelligence[0]
		if !intelligenceFeatures.Enabled.IsNull() && !intelligenceFeatures.Enabled.IsUnknown() {
			result.Intelligence.Enabled = intelligenceFeatures.Enabled.ValueBool()
		}
	}

	return result
}

// ParsePluginSDK converts the raw features list from the PluginSDK-based
// provider's configuration to the internal Features representation while
// handling default values.
//
// Only client-level features are parsed: resource-level features are
// consumed by Framework-based resources, which receive their features via
// Parse.
func ParsePluginSDK(raw []any) *Features {
	result := DefaultFeatures()
	if len(raw) == 0 {
		return result
	}
	features, ok := raw[0].(map[string]any)
	if !ok {
		return result
	}

	// parse client features
	if clientBlock, ok := features["client"].([]any); ok && len(clientBlock) > 0 {
		if clientFeatures, ok := clientBlock[0].(map[string]any); ok {
			if v, ok := clientFeatures["read_caching"].(bool); ok {
				result.Client.ReadCaching = v
			}
		}
	}

	return result
}
