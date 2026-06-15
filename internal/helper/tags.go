package helper

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
)

// TagsToMap converts a slice of client.Tag to a types.Map
// It returns a populated map if tags exist, or an empty map if no tags are present
func TagsToMap(ctx context.Context, tags []client.Tag) (types.Map, diag.Diagnostics) {
	tagMap := make(map[string]string, len(tags))
	for _, tag := range tags {
		tagMap[tag.Key] = tag.Value
	}
	return types.MapValueFrom(ctx, types.StringType, tagMap)
}

// MapToTags converts a map attribute of tags to a slice of client.Tag objects.
// If the map is empty, it returns an empty slice to clear any existing tags.
func MapToTags(ctx context.Context, tagsMap types.Map) ([]client.Tag, diag.Diagnostics) {
	tagMap := make(map[string]string, len(tagsMap.Elements()))
	if diags := tagsMap.ElementsAs(ctx, &tagMap, false); diags.HasError() {
		return nil, diags
	}

	tags := make([]client.Tag, 0, len(tagMap))
	for k, v := range tagMap {
		tags = append(tags, client.Tag{Key: k, Value: v})
	}

	return tags, nil
}

// MergeTags returns the effective set of tags for a resource: the provider's
// default tags overlaid with the resource's own tags, where the resource wins
// on a key collision. An unknown resource map yields an unknown result so the
// merge is deferred until the value is known.
func MergeTags(ctx context.Context, defaults map[string]string, resourceTags types.Map) (types.Map, diag.Diagnostics) {
	var diags diag.Diagnostics
	if resourceTags.IsUnknown() {
		return types.MapUnknown(types.StringType), diags
	}

	merged := make(map[string]string, len(defaults))
	for k, v := range defaults {
		merged[k] = v
	}
	if !resourceTags.IsNull() {
		var rt map[string]string
		diags.Append(resourceTags.ElementsAs(ctx, &rt, false)...)
		if diags.HasError() {
			return types.MapNull(types.StringType), diags
		}
		for k, v := range rt {
			merged[k] = v
		}
	}

	if len(merged) > client.MaxTagsPerResource {
		diags.AddError(
			"Too many tags",
			fmt.Sprintf("Merging default_tags with the resource's tags results in %d tags, which exceeds the maximum of %d per resource.",
				len(merged), client.MaxTagsPerResource),
		)
		return types.MapNull(types.StringType), diags
	}

	return types.MapValueFrom(ctx, types.StringType, merged)
}

// RemoveDefaultTags returns the resource-owned subset of a tag set by dropping
// any tag whose key and value match a provider default. It is the read-side
// inverse of MergeTags.
func RemoveDefaultTags(ctx context.Context, tags []client.Tag, defaults map[string]string) (types.Map, diag.Diagnostics) {
	owned := make(map[string]string, len(tags))
	for _, tag := range tags {
		if v, ok := defaults[tag.Key]; ok && v == tag.Value {
			continue
		}
		owned[tag.Key] = tag.Value
	}
	return types.MapValueFrom(ctx, types.StringType, owned)
}
