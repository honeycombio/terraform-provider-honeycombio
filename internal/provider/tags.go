package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper"
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

// tagsAllSchema returns the schema for the computed tags_all attribute: the
// effective set of tags after merging the provider's default_tags.
func tagsAllSchema() schema.MapAttribute {
	return schema.MapAttribute{
		Description: "The effective tags on the resource: the provider's `default_tags` merged with the resource's `tags` (resource tags win on a key collision).",
		Computed:    true,
		ElementType: types.StringType,
		PlanModifiers: []planmodifier.Map{
			mapplanmodifier.UseStateForUnknown(),
		},
	}
}

// modifyPlanForDefaultTags computes the planned tags_all value by merging the
// provider's default tags into the planned tags. It is shared by every resource
// that supports tags.
func modifyPlanForDefaultTags(ctx context.Context, defaults map[string]string, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}

	var tags types.Map
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("tags"), &tags)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tagsAll, diags := helper.MergeTags(ctx, defaults, tags)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("tags_all"), tagsAll)...)
}
