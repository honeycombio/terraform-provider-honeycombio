package features

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	pluginSchema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// GetFeaturesBlock returns the features schema for plugin-based providers
func GetFeaturesBlock() schema.Block {
	return schema.ListNestedBlock{
		MarkdownDescription: "Provider features to enable experimental or opt-in behaviors.",
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
								MarkdownDescription: "Enable import-on-conflict behavior for column resources.",
								Optional:            true,
							},
						},
					},
				},
			},
		},
	}
}

// GetSDKFeaturesSchema returns the features schema for SDK-based providers
func GetSDKFeaturesSchema() *pluginSchema.Schema {
	return &pluginSchema.Schema{
		Type:        pluginSchema.TypeList,
		Optional:    true,
		MaxItems:    1,
		Description: "Provider features to enable experimental or opt-in behaviors.",
		Elem: &pluginSchema.Resource{
			Schema: map[string]*pluginSchema.Schema{
				"column": {
					Type:        pluginSchema.TypeList,
					Optional:    true,
					MaxItems:    1,
					Description: "Column resource features.",
					Elem: &pluginSchema.Resource{
						Schema: map[string]*pluginSchema.Schema{
							"import_on_conflict": {
								Type:        pluginSchema.TypeBool,
								Optional:    true,
								Description: "Enable import-on-conflict behavior for column resources.",
							},
						},
					},
				},
			},
		},
	}
}
