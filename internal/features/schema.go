package features

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	pluginsdk "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// GetFeaturesBlock returns the schema for the features block for the Framework-based provider.
func GetFeaturesBlock() schema.Block {
	return schema.ListNestedBlock{
		MarkdownDescription: "The features block allows customization of the behavior of the Honeycomb Provider.",
		Validators: []validator.List{
			listvalidator.SizeBetween(1, 1),
		},
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"column": schema.ListNestedBlock{
					MarkdownDescription: "Column resource features.",
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"import_on_conflict": schema.BoolAttribute{
								MarkdownDescription: "This changes the creation behavior of the column resource to import an existing column if it already exists, rather than erroring out.",
								Optional:            true,
							},
						},
					},
				},
				"dataset": schema.ListNestedBlock{
					MarkdownDescription: "Dataset resource features.",
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"import_on_conflict": schema.BoolAttribute{
								MarkdownDescription: "This changes the creation behavior of the dataset resource to import an existing dataset if it already exists, rather than erroring out.",
								Optional:            true,
							},
						},
					},
				},
			},
		},
	}
}

// GetPluginSDKFeaturesSchema returns the schema for the features block for the PluginSDK-based provider.
func GetPluginSDKFeaturesSchema() *pluginsdk.Schema {
	return &pluginsdk.Schema{
		Type:        pluginsdk.TypeList,
		Optional:    true,
		MaxItems:    1,
		Description: "The features block allows customization of the behavior of the Honeycomb Provider.",
		Elem: &pluginsdk.Resource{
			Schema: map[string]*pluginsdk.Schema{
				"column": {
					Type:        pluginsdk.TypeList,
					Optional:    true,
					MaxItems:    1,
					Description: "Column resource features.",
					Elem: &pluginsdk.Resource{
						Schema: map[string]*pluginsdk.Schema{
							"import_on_conflict": {
								Type:        pluginsdk.TypeBool,
								Optional:    true,
								Description: "This changes the creation behavior of the column resource to import an existing column if it already exists, rather than erroring out.",
							},
						},
					},
				},
				"dataset": {
					Type:        pluginsdk.TypeList,
					Optional:    true,
					MaxItems:    1,
					Description: "Dataset resource features.",
					Elem: &pluginsdk.Resource{
						Schema: map[string]*pluginsdk.Schema{
							"import_on_conflict": {
								Type:        pluginsdk.TypeBool,
								Optional:    true,
								Description: "This changes the creation behavior of the dataset resource to import an existing dataset if it already exists, rather than erroring out.",
							},
						},
					},
				},
			},
		},
	}
}
