package v2

import (
	"context"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	hnyclient "github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/test"
)

func TestClient_APIKeys(t *testing.T) {
	ctx := context.Background()
	c := newTestClient(t)
	env := newTestEnvironment(ctx, t, c)

	// create a new key
	keyName := test.RandomStringWithPrefix("test.", 10)
	k, err := c.APIKeys.Create(ctx, &APIKey{
		Name:    helper.ToPtr(keyName),
		KeyType: "ingest",
		Environment: &Environment{
			ID: env.ID,
		},
		Permissions: &APIKeyPermissions{
			CreateDatasets: true,
		},
	})
	require.NoError(t, err)
	assert.NotEmpty(t, k.ID)
	assert.NotEmpty(t, k.Secret)
	assert.Equal(t, keyName, *k.Name)
	assert.False(t, *k.Disabled)
	assert.True(t, k.Permissions.CreateDatasets)

	// read the key back and compare
	key, err := c.APIKeys.Get(ctx, k.ID)
	require.NoError(t, err)
	assert.Equal(t, k.ID, key.ID)
	assert.Equal(t, k.Name, key.Name)
	assert.Equal(t, k.KeyType, key.KeyType)
	assert.Equal(t, k.Disabled, key.Disabled)
	assert.Equal(t, k.Environment.ID, key.Environment.ID)
	assert.Equal(t, k.Permissions.CreateDatasets, key.Permissions.CreateDatasets)
	assert.WithinDuration(t, k.Timestamps.CreatedAt, key.Timestamps.CreatedAt, 5*time.Second)
	assert.WithinDuration(t, k.Timestamps.UpdatedAt, key.Timestamps.UpdatedAt, 5*time.Second)

	// update the key's name and disable it
	keyName = test.RandomStringWithPrefix("test.", 10)
	key, err = c.APIKeys.Update(ctx, &APIKey{
		ID:       k.ID,
		Name:     helper.ToPtr(keyName),
		Disabled: helper.ToPtr(true),
	})
	require.NoError(t, err)
	assert.Equal(t, k.ID, key.ID)
	assert.Equal(t, keyName, *key.Name)
	assert.True(t, *key.Disabled)
	assert.WithinDuration(t, time.Now(), key.Timestamps.UpdatedAt, 5*time.Second)

	// delete the key
	err = c.APIKeys.Delete(ctx, k.ID)
	require.NoError(t, err)

	// verify that the key was deleted
	_, err = c.APIKeys.Get(ctx, k.ID)
	var de hnyclient.DetailedError
	require.ErrorAs(t, err, &de)
	assert.True(t, de.IsNotFound())
}

func TestClient_APIKeys_Pagination(t *testing.T) {
	ctx := context.Background()
	c := newTestClient(t)
	env := newTestEnvironment(ctx, t, c)

	// create a bunch of keys
	numKeys := int(math.Floor(1.5 * float64(defaultPageSize)))
	testKeys := make([]*APIKey, numKeys)
	for i := range numKeys {
		k, err := c.APIKeys.Create(ctx, &APIKey{
			Name:    helper.ToPtr(fmt.Sprintf("test.%d", i)),
			KeyType: "ingest",
			Environment: &Environment{
				ID: env.ID,
			},
		})
		require.NoError(t, err)
		testKeys[i] = k
	}
	t.Cleanup(func() {
		for _, k := range testKeys {
			c.APIKeys.Delete(ctx, k.ID)
		}
	})

	t.Run("happy path", func(t *testing.T) {
		keys := make([]*APIKey, 0)
		pager, err := c.APIKeys.List(ctx)
		require.NoError(t, err)

		items, err := pager.Next(ctx)
		require.NoError(t, err)
		assert.Len(t, items, defaultPageSize, "incorrect number of items")
		assert.True(t, pager.HasNext(), "should have more pages")
		keys = append(keys, items...)

		for pager.HasNext() {
			items, err = pager.Next(ctx)
			require.NoError(t, err)
			keys = append(keys, items...)
		}
		// we can't guarantee that there are exactly numKeys keys, but there should be at least that many
		assert.GreaterOrEqual(t, len(keys), numKeys, "should have at least %d keys", numKeys)
	})

	t.Run("works with custom page size", func(t *testing.T) {
		pageSize := 5
		pager, err := c.APIKeys.List(ctx, PageSize(pageSize))
		require.NoError(t, err)

		items, err := pager.Next(ctx)
		require.NoError(t, err)
		assert.Len(t, items, pageSize, "incorrect number of items")
		assert.True(t, pager.HasNext(), "should have more pages")
	})
}
