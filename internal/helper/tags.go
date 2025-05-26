package helper

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/honeycombio/terraform-provider-honeycombio/client"
)

// TagsToMap converts a slice of client.Tag to a types.Map
// It returns a populated map if tags exist, or a null map if no tags are present
func TagsToMap(ctx context.Context, tags []client.Tag) (types.Map, diag.Diagnostics) {
	if len(tags) > 0 {
		tagMap := make(map[string]string)
		for _, tag := range tags {
			tagMap[tag.Key] = tag.Value
		}
		return types.MapValueFrom(ctx, types.StringType, tagMap)
	}
	return types.MapNull(types.StringType), nil
}

// MapToTags converts a map attribute of tags to a slice of client.Tag objects.
// If the map is null, it returns an empty slice to clear any existing tags.
func MapToTags(ctx context.Context, tagsMap types.Map) ([]client.Tag, diag.Diagnostics) {
	var tags []client.Tag

	if !tagsMap.IsNull() {
		var tagMap map[string]string
		diags := tagsMap.ElementsAs(ctx, &tagMap, false)
		if diags.HasError() {
			return nil, diags
		}

		for k, v := range tagMap {
			tags = append(tags, client.Tag{Key: k, Value: v})
		}
	} else {
		// if 'tags' is not present in the config, set to empty slice
		// to clear the tags
		tags = make([]client.Tag, 0)
	}

	return tags, nil
}
