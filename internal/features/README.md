# Features Package

This package provides centralized management of provider-level features for the Honeycomb Terraform provider. It supports both the Plugin Framework provider (`internal/provider`) and the SDK-based provider (`honeycombio`).

## Overview

The features package contains:
- Type definitions for features and their Terraform schema models
- Schema generation functions for both PluginFramework and SDK providers
- Parsing logic to convert provider configurations to internal representations
- Default value management and configuration utilities
- Comprehensive test coverage for all features scenarios

## Architecture

### Dual Provider Support

The package supports two different Terraform provider frameworks:

1. **Framework Provider** (`internal/provider/provider.go`): Plugin framework
2. **SDK Provider** (`honeycombio/provider.go`): SDK v2

Both providers share the same features logic but use different parsing and schema approaches.

## Usage

### Provider Configuration

Users can enable features in their Terraform configuration with identical syntax across both providers:

```hcl
provider "honeycombio" {
  features {
    column {
      import_on_conflict = true
    }
  }
}
```

### Accessing Features in Framework Resources

Framework-based resources access features through the `ConfiguredClient`:

```go
func (r *myResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    client := getClientFromResourceRequest(&req.ProviderData)
    
    // Access features
    if client.Features.Column.ImportOnConflict {
        // Implement import-on-conflict behavior
    }
}
```

### Accessing Features in SDK Resources

SDK-based resources access features through the `SDKConfiguredClient`:

```go
func resourceMyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
    configuredClient, err := getSDKConfiguredClient(meta)
    if err != nil {
        return diagFromErr(err)
    }
    
    client := configuredClient.Client
    features := configuredClient.Features
    
    // Access features
    if features.Column.ImportOnConflict {
        // Implement import-on-conflict behavior
    }
}
```

## Features

### Current Features

#### Column Features (`column`)

- **`import_on_conflict`** (default: `false`): When enabled, allows column resources to import existing columns instead of failing when a column with the same name already exists.

### Default Values

All features have safe defaults defined in `defaults.go`:

```go
func DefaultFeatures() Features {
    return Features{
        Column: FeaturesColumn{
            ImportOnConflict: false,
        },
    }
}
```

### Practical Example: Column Import on Conflict

When `import_on_conflict = true`, the column resource behavior changes:

```go
// In honeycombio/resource_column.go
if features != nil && features.Column.ImportOnConflict {
    existing, err := client.Columns.GetByKeyName(ctx, dataset, column.KeyName)
    if err == nil {
        // Import existing column instead of failing
        d.SetId(existing.ID)
        return resourceColumnUpdate(ctx, d, meta)
    }
}
```

## Adding New Features

To add new features, follow these steps:

### 1. Define Features Types

In `features.go`, add your new features types:

```go
// Add to Features struct
type Features struct {
    Column  FeaturesColumn
    Dataset FeaturesDataset  // New feature category
}

// Add new features struct
type FeaturesDataset struct {
    AutoRetention bool
    EnableCaching bool
}

// Add Framework model struct
type FeaturesDatasetModel struct {
    AutoRetention types.Bool `tfsdk:"auto_retention"`
    EnableCaching types.Bool `tfsdk:"enable_caching"`
}

// Add to Model struct
type Model struct {
    Column  []FeaturesColumnModel
    Dataset []FeaturesDatasetModel  // New field
}
```

### 2. Update Parsing Functions

Extend parsing functions in `features.go` depending on the provider type:

```go
// Update Parse function for Framework provider
func Parse(m Model) *Features {
    features := DefaultFeatures()
    
    // ... existing column parsing ...
    
    // Parse dataset features
    if len(m.Dataset) > 0 {
        datasetFeatures := m.Dataset[0]
        if !datasetFeatures.AutoRetention.IsNull() && !datasetFeatures.AutoRetention.IsUnknown() {
            features.Dataset.AutoRetention = datasetFeatures.AutoRetention.ValueBool()
        }
        // ... similar for other dataset features
    }
    
    return &features
}

// Update ParseSDKResourceData function for SDK provider
func ParseSDKResourceData(d interface{}) *Features {
    // ... existing parsing logic ...
    
    // Parse dataset features for SDK provider
    if datasetRaw, ok := featuresMap["dataset"]; ok {
        datasetList := datasetRaw.([]interface{})
        if len(datasetList) > 0 {
            datasetMap := datasetList[0].(map[string]interface{})
            if autoRetention, ok := datasetMap["auto_retention"]; ok {
                features.Dataset.AutoRetention = autoRetention.(bool)
            }
        }
    }
}
```

### 3. Update Schema Functions

In `schema.go`, add schema definitions for applicable provider type:

```go
// Update Framework schema
func GetFeaturesBlock() schema.Block {
    return schema.ListNestedBlock{
        // ... existing config ...
        NestedObject: schema.NestedBlockObject{
            Blocks: map[string]schema.Block{
                "column":  /* existing column schema */,
                "dataset": getDatasetFeatureBlock(),  // New block
            },
        },
    }
}

// Update SDK schema
func GetSDKFeaturesSchema() *pluginSchema.Schema {
    return &pluginSchema.Schema{
        // ... existing config ...
        Elem: &pluginSchema.Resource{
            Schema: map[string]*pluginSchema.Schema{
                "column":  /* existing column schema */,
                "dataset": getSDKDatasetFeatureSchema(),  // New schema
            },
        },
    }
}
```

### 4. Update Defaults

In `defaults.go`, add default values:

```go
func DefaultFeatures() Features {
    return Features{
        Column:  defaultColumnFeatures(),
        Dataset: defaultDatasetFeatures(),  // New defaults
    }
}

func defaultDatasetFeatures() FeaturesDataset {
    return FeaturesDataset{
        AutoRetention: false,
        EnableCaching: true,
    }
}
```

### 5. Add Tests

Create comprehensive tests in `features_test.go`:

```go
func TestDatasetFeatureParsing(t *testing.T) {
    // Test Framework parsing
    model := Model{
        Dataset: []FeaturesDatasetModel{
            {
                AutoRetention: types.BoolValue(true),
                EnableCaching: types.BoolValue(false),
            },
        },
    }
    
    features := Parse(model)
    assert.True(t, features.Dataset.AutoRetention)
    assert.False(t, features.Dataset.EnableCaching)
}
```

## Schema Consistency

Both providers generate identical configuration schemas through careful coordination:

- **Descriptions**: Exact same text across both schema functions
- **Structure**: Identical nesting and field names  
- **Types**: Consistent data types and validation rules
- **Defaults**: Shared default values from `defaults.go`

This ensures users have the same experience regardless of which provider implementation is active.