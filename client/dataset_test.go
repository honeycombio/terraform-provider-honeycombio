package client_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/test"
)

func TestDatasets(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	c := newTestClient(t)

	t.Run("happy path", func(t *testing.T) {
		// create a new dataset
		name := test.RandomStringWithPrefix("test.", 10)
		description := test.RandomString(70)
		ds, err := c.Datasets.Create(ctx, &client.Dataset{
			Name:            name,
			Description:     description,
			ExpandJSONDepth: 2,
		})
		require.NoError(t, err)
		assert.Equal(t, name, ds.Name)
		assert.NotEmpty(t, ds.Slug)
		assert.Equal(t, description, ds.Description)
		assert.Equal(t, 2, ds.ExpandJSONDepth)
		assert.True(t, *ds.Settings.DeleteProtected)
		assert.WithinDuration(t, time.Now(), ds.CreatedAt, 5*time.Second)

		// read back the dataset by Slug and compare
		dataset, err := c.Datasets.Get(ctx, ds.Slug)
		require.NoError(t, err)
		assert.Equal(t, ds.Name, dataset.Name)
		assert.Equal(t, ds.Slug, dataset.Slug)
		assert.Equal(t, ds.Description, dataset.Description)
		assert.Equal(t, ds.ExpandJSONDepth, dataset.ExpandJSONDepth)
		if assert.NotNil(t, dataset.Settings) {
			assert.True(t, *dataset.Settings.DeleteProtected)
		}

		// list all datasets and verify the dataset is in the list
		dds, err := c.Datasets.List(ctx)
		require.NoError(t, err)
		assert.Contains(t, dds, *dataset)

		// update the dataset's description and expand_json_depth
		newDescription := test.RandomString(70)
		dataset, err = c.Datasets.Update(ctx, &client.Dataset{
			Slug:            dataset.Slug,
			Description:     newDescription,
			ExpandJSONDepth: 5,
			Settings:        dataset.Settings,
		})
		require.NoError(t, err)
		assert.Equal(t, ds.Slug, dataset.Slug)
		assert.Equal(t, newDescription, dataset.Description)
		assert.Equal(t, 5, dataset.ExpandJSONDepth)
		assert.True(t, *dataset.Settings.DeleteProtected)

		// try to delete the dataset with deletion protection enabled
		var de client.DetailedError
		err = c.Datasets.Delete(ctx, dataset.Slug)
		require.ErrorAs(t, err, &de)
		assert.Equal(t, http.StatusConflict, de.Status)

		// disable deletion protection and delete the dataset
		_, err = c.Datasets.Update(ctx, &client.Dataset{
			Slug: dataset.Slug,
			Settings: client.DatasetSettings{
				DeleteProtected: helper.ToPtr(false),
			},
		})
		require.NoError(t, err)
		err = c.Datasets.Delete(ctx, dataset.Slug)
		require.NoError(t, err)

		// verify the dataset was deleted
		_, err = c.Datasets.Get(ctx, dataset.Slug)
		require.ErrorAs(t, err, &de)
		assert.True(t, de.IsNotFound())
	})

	t.Run("returns DatasetExistsErr when creating dataset with same name", func(t *testing.T) {
		name := test.RandomStringWithPrefix("test.", 10)
		ds, err := c.Datasets.Create(ctx, &client.Dataset{
			Name: name,
		})
		require.NoError(t, err)
		t.Cleanup(func() {
			_, _ = c.Datasets.Update(ctx, &client.Dataset{
				Slug: ds.Slug,
				Settings: client.DatasetSettings{
					DeleteProtected: helper.ToPtr(false),
				},
			})
			_ = c.Datasets.Delete(ctx, ds.Slug)
		})

		_, err = c.Datasets.Create(ctx, &client.Dataset{
			Name: name,
		})
		require.ErrorIs(t, err, client.ErrDatasetExists)
	})
}
