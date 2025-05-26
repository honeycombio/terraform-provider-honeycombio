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
