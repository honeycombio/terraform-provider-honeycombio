package provider

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/validation"
)

func detailFilterSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "Attributes to filter the results with. `name` must be set when providing a filter.",
		Validators:  []validator.List{listvalidator.SizeAtMost(1)},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					Required:    true,
					Description: "The name of the detail field to filter by. Currently only 'name' is supported.",
					Validators:  []validator.String{stringvalidator.OneOf("name")},
				},
				"value": schema.StringAttribute{
					Optional:    true,
					Description: "The value of the detail field to match on.",
					Validators: []validator.String{
						stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("value_regex")),
					},
				},
				"value_regex": schema.StringAttribute{
					Optional:    true,
					Description: "A regular expression string to apply to the value of the detail field to match on.",
					Validators: []validator.String{
						stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("value")),
						validation.IsValidRegExp(),
					},
				},
			},
		},
	}
}
