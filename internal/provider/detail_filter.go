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
		MarkdownDescription: "Attributes to filter the results with. Multiple `detail_filter` blocks can be provided, and all conditions must be satisfied (AND logic).",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					Required:            true,
					MarkdownDescription: "The name of the detail field to filter by. This must match a schema attribute of the resource (e.g., `name`, `description`, `id`).",
				},
				"operator": schema.StringAttribute{
					Optional: true,
					MarkdownDescription: `The comparison operator to use for filtering. Defaults to ` + "`equals`" + `. Valid operators include:
  * ` + "`equals`" + `, ` + "`=`" + `, ` + "`eq`" + ` - Exact match comparison
  * ` + "`not-equals`" + `, ` + "`!=`" + `, ` + "`ne`" + ` - Inverse exact match comparison
  * ` + "`contains`" + `, ` + "`in`" + ` - Substring inclusion check
  * ` + "`does-not-contain`" + `, ` + "`not-in`" + ` - Inverse substring inclusion check
  * ` + "`starts-with`" + ` - Prefix matching
  * ` + "`does-not-start-with`" + ` - Inverse prefix matching
  * ` + "`ends-with`" + ` - Suffix matching
  * ` + "`does-not-end-with`" + ` - Inverse suffix matching
  * ` + "`>`" + `, ` + "`gt`" + ` - Numeric greater than comparison
  * ` + "`>=`" + `, ` + "`ge`" + ` - Numeric greater than or equal comparison
  * ` + "`<`" + `, ` + "`lt`" + ` - Numeric less than comparison
  * ` + "`<=`" + `, ` + "`le`" + ` - Numeric less than or equal comparison
  * ` + "`does-not-exist`" + ` - Field absence check`,
					Validators: []validator.String{
						stringvalidator.OneOf(
							"equals", "=", "eq", "not-equals", "!=", "ne", "contains", "in", "does-not-contain",
							"not-in", "starts-with", "does-not-start-with", "ends-with", "does-not-end-with",
							">", "gt", ">=", "ge", "<", "lt", "<=", "le", "does-not-exist", "does-not-exist",
						),
					},
				},
				"value": schema.StringAttribute{
					Optional:            true,
					MarkdownDescription: "The value of the detail field to match on. Required unless `value_regex` is set or `operator` is `does-not-exist`.",
					Validators: []validator.String{
						stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("value_regex")),
					},
				},
				"value_regex": schema.StringAttribute{
					Optional:            true,
					MarkdownDescription: "A regular expression string to apply to the value of the detail field to match on. Required unless `value` is set or `operator` is `does-not-exist`.",
					Validators: []validator.String{
						stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("value")),
						validation.IsValidRegExp(),
					},
				},
			},
		},
	}
}
