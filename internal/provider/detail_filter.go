package provider

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/validation"
)

func detailFilterSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "Attributes to filter the results with. Multiple filters can be specified, and all conditions must be satisfied (AND logic).",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					Required:    true,
					Description: "The field to filter by (e.g., 'name', 'id', 'description', etc.).",
				},
				"operator": schema.StringAttribute{
					Optional:    true,
					Description: "The comparison operator. The default is 'equals'.",
					Validators: []validator.String{
						stringvalidator.OneOf(
							"equals", "=", "eq", "not-equals", "!=", "ne", "contains", "in", "does-not-contain",
							"not-in", "starts-with", "does-not-start-with", "ends-with", "does-not-end-with",
							">", "gt", ">=", "ge", "<", "lt", "<=", "le", "does-not-exist", "does-not-exist",
						),
					},
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
