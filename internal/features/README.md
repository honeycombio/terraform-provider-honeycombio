# Features Package

The features package is a simple way of providing resource-level "feature toggles" so users of the Honeycomb Provider can modify behavior.

The package only supports Framework-based resources: if the resource you are looking to add a feature toggle for is still based on the PluginSDK please migrate the resource to the Framework as a first step.

## Adding a new Feature

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

Extend parsing functions in `features.go`:

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
```

### 3. Update Schema Functions

As we are bridging (via a MuxServer) a PluginSDK-based and a Framework-based provider, the two of them need to start up with _identical_ configurations or they will panic (we have tests to catch this -- don't worry!).
So, even though the SDKv2-based provider doesn't make use of features, it needs to have the same provider schema.

In `schema.go`, add schema definitions for applicable provider type:

```go
// Update Framework schema
func GetFeaturesBlock() schema.Block {
    return schema.ListNestedBlock{
        // ... existing config ...
        NestedObject: schema.NestedBlockObject{
            Blocks: map[string]schema.Block{
                "column":  /* existing column schema */,
                "dataset": schema.ListNestedBlock{...}
            },
        },
    }
}

// Update PluginSDK schema
func GetPluginSDKFeaturesSchema() *pluginsdk.Schema {
    return &pluginsdk.Schema{
        // ... existing config ...
        Elem: &pluginsdk.Resource{
            Schema: map[string]*pluginsdk.Schema{
                "column":  /* existing column schema */,
                "dataset": { ... }
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
