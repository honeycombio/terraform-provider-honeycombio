package helper

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
)

func tagsMap(t *testing.T, m map[string]string) types.Map {
	t.Helper()
	v, diags := types.MapValueFrom(context.Background(), types.StringType, m)
	require.False(t, diags.HasError())
	return v
}

func TestMergeTags(t *testing.T) {
	ctx := context.Background()

	t.Run("resource tags win over defaults", func(t *testing.T) {
		got, diags := MergeTags(ctx,
			map[string]string{"team": "sre", "env": "prod"},
			tagsMap(t, map[string]string{"team": "platform"}),
		)
		require.False(t, diags.HasError())

		var out map[string]string
		got.ElementsAs(ctx, &out, false)
		assert.Equal(t, map[string]string{"team": "platform", "env": "prod"}, out)
	})

	t.Run("no defaults returns the resource tags", func(t *testing.T) {
		got, diags := MergeTags(ctx, nil, tagsMap(t, map[string]string{"team": "platform"}))
		require.False(t, diags.HasError())

		var out map[string]string
		got.ElementsAs(ctx, &out, false)
		assert.Equal(t, map[string]string{"team": "platform"}, out)
	})

	t.Run("null resource tags returns the defaults", func(t *testing.T) {
		got, diags := MergeTags(ctx, map[string]string{"team": "sre"}, types.MapNull(types.StringType))
		require.False(t, diags.HasError())

		var out map[string]string
		got.ElementsAs(ctx, &out, false)
		assert.Equal(t, map[string]string{"team": "sre"}, out)
	})

	t.Run("unknown resource tags returns unknown", func(t *testing.T) {
		got, diags := MergeTags(ctx, map[string]string{"team": "sre"}, types.MapUnknown(types.StringType))
		require.False(t, diags.HasError())
		assert.True(t, got.IsUnknown())
	})

	t.Run("errors when the merged set exceeds the limit", func(t *testing.T) {
		defaults := make(map[string]string, client.MaxTagsPerResource)
		for i := 0; i < client.MaxTagsPerResource; i++ {
			defaults[fmt.Sprintf("d%d", i)] = "v"
		}
		_, diags := MergeTags(ctx, defaults, tagsMap(t, map[string]string{"extra": "v"}))
		assert.True(t, diags.HasError())
	})
}

func TestRemoveDefaultTags(t *testing.T) {
	ctx := context.Background()

	t.Run("drops tags matching a default key and value", func(t *testing.T) {
		all := []client.Tag{{Key: "team", Value: "platform"}, {Key: "env", Value: "prod"}}
		got, diags := RemoveDefaultTags(ctx, all, map[string]string{"team": "sre", "env": "prod"})
		require.False(t, diags.HasError())

		var out map[string]string
		got.ElementsAs(ctx, &out, false)
		assert.Equal(t, map[string]string{"team": "platform"}, out)
	})

	t.Run("keeps a tag whose value differs from the default", func(t *testing.T) {
		all := []client.Tag{{Key: "team", Value: "platform"}}
		got, diags := RemoveDefaultTags(ctx, all, map[string]string{"team": "sre"})
		require.False(t, diags.HasError())

		var out map[string]string
		got.ElementsAs(ctx, &out, false)
		assert.Equal(t, map[string]string{"team": "platform"}, out)
	})

	t.Run("no defaults returns all tags", func(t *testing.T) {
		all := []client.Tag{{Key: "team", Value: "platform"}}
		got, diags := RemoveDefaultTags(ctx, all, nil)
		require.False(t, diags.HasError())

		var out map[string]string
		got.ElementsAs(ctx, &out, false)
		assert.Equal(t, map[string]string{"team": "platform"}, out)
	})
}
