package client_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthMetadata(t *testing.T) {
	t.Parallel()

	c := newTestClient(t)

	t.Run("List", func(t *testing.T) {
		ctx := t.Context()
		metadata, err := c.Auth.List(ctx)
		require.NoError(t, err)

		assert.NotEmpty(t, metadata.APIKeyAccess)
		assert.NotEmpty(t, metadata.Team.Name)
		assert.NotEmpty(t, metadata.Team.Slug)
		// classic environment's don't have environment name and slugs
		if c.IsClassic(ctx) {
			assert.Empty(t, metadata.Environment.Name)
			assert.Empty(t, metadata.Environment.Slug)
		} else {
			assert.NotEmpty(t, metadata.Environment.Name)
			assert.NotEmpty(t, metadata.Environment.Slug)
		}
	})
}
