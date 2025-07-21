package provider

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/modifiers"
)

// tagsSchema returns a schema for tags that can be used in resources.
func tagsSchema() schema.MapAttribute {
	return schema.MapAttribute{
		Description: "A map of tags to assign to the resource.",
		Optional:    true,
		Computed:    true,
		ElementType: types.StringType,
		Default: mapdefault.StaticValue( // Default to an empty map
			types.MapValueMust(
				types.StringType,
				map[string]attr.Value{},
			),
		),
		PlanModifiers: []planmodifier.Map{
			modifiers.EquivalentTags(),
		},
		Validators: []validator.Map{
			mapvalidator.SizeAtMost(client.MaxTagsPerResource),
			mapvalidator.KeysAre(
				stringvalidator.RegexMatches(
					client.TagKeyValidationRegex,
					"must only contain lowercase letters, and be 1-32 characters long",
				),
			),
			mapvalidator.ValueStringsAre(
				stringvalidator.RegexMatches(
					client.TagValueValidationRegex,
					"must begin with a lowercase letter, be between 1-128 characters long, and only contain lowercase alphanumeric characters, -, or /",
				),
			),
		},
	}
}
