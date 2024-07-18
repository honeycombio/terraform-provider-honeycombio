package v2

import (
	"context"
	"math"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	hnyclient "github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/test"
)

func TestClient_Environments(t *testing.T) {
	ctx := context.Background()
	c := newTestClient(t)

	t.Run("happy path", func(t *testing.T) {
		// create a new Environment
		newEnv := &Environment{
			Name:        test.RandomStringWithPrefix("test.", 20),
			Description: helper.ToPtr(test.RandomString(50)),
			Color:       helper.ToPtr(EnvironmentColorBlue),
		}
		e, err := c.Environments.Create(ctx, newEnv)
		require.NoError(t, err)
		assert.NotEmpty(t, e.ID)
		assert.Equal(t, newEnv.Name, e.Name)
		assert.NotEmpty(t, e.Slug)
		assert.Equal(t, newEnv.Description, e.Description)
		assert.Equal(t, newEnv.Color, e.Color)
		if assert.NotNil(t, e.Settings) {
			assert.True(t, *e.Settings.DeleteProtected)
		}

		// read the Environment back and compare
		env, err := c.Environments.Get(ctx, e.ID)
		require.NoError(t, err)
		assert.Equal(t, e.ID, env.ID)
		assert.Equal(t, e.Name, env.Name)
		assert.Equal(t, e.Slug, env.Slug)
		assert.Equal(t, e.Description, env.Description)
		assert.Equal(t, e.Color, env.Color)
		if assert.NotNil(t, env.Settings) {
			assert.True(t, *env.Settings.DeleteProtected)
		}

		// update the Environment's description and color
		newDescription := helper.ToPtr(test.RandomString(50))
		env.Description = newDescription
		env.Color = helper.ToPtr(EnvironmentColorGreen)
		env, err = c.Environments.Update(ctx, env)
		require.NoError(t, err)
		assert.Equal(t, e.ID, env.ID)
		assert.Equal(t, newDescription, env.Description)
		assert.Equal(t, EnvironmentColorGreen, *env.Color)

		// try to delete the environment with delete protection enabled
		var de hnyclient.DetailedError
		err = c.Environments.Delete(ctx, env.ID)
		require.ErrorAs(t, err, &de)
		assert.Equal(t, http.StatusConflict, de.Status)

		// disable deletion protection and delete the Environment
		_, err = c.Environments.Update(ctx, &Environment{
			ID: env.ID,
			Settings: &EnvironmentSettings{
				DeleteProtected: helper.ToPtr(false),
			},
		})
		require.NoError(t, err)
		err = c.Environments.Delete(ctx, env.ID)
		require.NoError(t, err)

		// verify the Environment was deleted
		_, err = c.Environments.Get(ctx, env.ID)
		require.ErrorAs(t, err, &de)
		assert.True(t, de.IsNotFound())
	})
}

func TestClient_Environments_Pagination(t *testing.T) {
	ctx := context.Background()
	c := newTestClient(t)

	// create a bunch of environments
	numEnvs := int(math.Floor(1.5 * float64(defaultPageSize)))
	testEnvs := make([]*Environment, numEnvs)
	for i := 0; i < numEnvs; i++ {
		e, err := c.Environments.Create(ctx, &Environment{
			Name: test.RandomStringWithPrefix("test.", 20),
		})
		require.NoError(t, err)
		testEnvs[i] = e
	}
	t.Cleanup(func() {
		for _, e := range testEnvs {
			c.Environments.Update(ctx, &Environment{
				ID: e.ID,
				Settings: &EnvironmentSettings{
					DeleteProtected: helper.ToPtr(false),
				},
			})
			c.Environments.Delete(ctx, e.ID)
		}
	})

	t.Run("happy path", func(t *testing.T) {
		envs := make([]*Environment, 0)
		pager, err := c.Environments.List(ctx)
		require.NoError(t, err)

		items, err := pager.Next(ctx)
		require.NoError(t, err)
		assert.Len(t, items, defaultPageSize, "incorrect number of items")
		assert.True(t, pager.HasNext(), "should have more pages")
		envs = append(envs, items...)

		for pager.HasNext() {
			items, err = pager.Next(ctx)
			require.NoError(t, err)
			envs = append(envs, items...)
		}
		// we can't guarantee that there are exactly numEnvs environments
		assert.GreaterOrEqual(t, len(envs), numEnvs, "should have at least %d environments", numEnvs)
	})

	t.Run("works with custom page size", func(t *testing.T) {
		pageSize := 5
		pager, err := c.Environments.List(ctx, PageSize(pageSize))
		require.NoError(t, err)

		items, err := pager.Next(ctx)
		require.NoError(t, err)
		assert.Len(t, items, pageSize, "incorrect number of items")
		assert.True(t, pager.HasNext(), "should have more pages")
	})
}
