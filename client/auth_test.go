package client_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthMetadata(t *testing.T) {
	t.Parallel()

	c := newTestClient(t)

	t.Run("List", func(t *testing.T) {
		ctx := context.Background()
		metadata, err := c.Auth.List(ctx)

		assert.NoError(t, err)
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
