package helper

import (
	"context"

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
