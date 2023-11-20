package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthMetadata(t *testing.T) {
	t.Parallel()

	c := newTestClient(t)

	t.Run("List", func(t *testing.T) {
		metadata, err := c.Auth.List(context.Background())

		assert.NoError(t, err)
		assert.NotEmpty(t, metadata.APIKeyAccess)
		assert.NotEmpty(t, metadata.Team.Name)
		assert.NotEmpty(t, metadata.Team.Slug)
		// classic environment's don't have environment name and slugs
		if len(c.apiKey) == 32 {
			assert.Empty(t, metadata.Environment.Name)
			assert.Empty(t, metadata.Environment.Slug)
		} else {
			assert.NotEmpty(t, metadata.Environment.Name)
			assert.NotEmpty(t, metadata.Environment.Slug)
		}
	})
}
