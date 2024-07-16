package client_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/client/errors"
)

func TestDatasets(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	c := newTestClient(t)
	datasetName := testDataset(t)

	currentDataset := &client.Dataset{
		Name: datasetName,
	}

	t.Run("List", func(t *testing.T) {
		d, err := c.Datasets.List(ctx)
		require.NoError(t, err)

		for _, dataset := range d {
			assert.NotNil(t, dataset.LastWrittenAt, "last written at is empty")
			assert.NotNil(t, dataset.CreatedAt, "created at is empty")
			// copy dynamic fields before asserting - will be skipped if expected dataset not found
			if dataset.Name == currentDataset.Name {
				currentDataset.Slug = dataset.Slug
				currentDataset.LastWrittenAt = dataset.LastWrittenAt
				currentDataset.CreatedAt = dataset.CreatedAt
			}
		}
		assert.Contains(t, d, *currentDataset, "could not find current dataset with List")
	})

	t.Run("Get", func(t *testing.T) {
		d, err := c.Datasets.Get(ctx, currentDataset.Slug)
		require.NoError(t, err)

		assert.NotNil(t, d.LastWrittenAt, "last written at is empty")
		assert.NotNil(t, d.CreatedAt, "created at is empty")
		// copy dynamic fields before asserting equality
		d.LastWrittenAt = currentDataset.LastWrittenAt
		d.CreatedAt = currentDataset.CreatedAt

		assert.Equal(t, *currentDataset, *d)
	})

	t.Run("Fail to Get bogus Dataset", func(t *testing.T) {
		_, err := c.Datasets.Get(ctx, "does-not-exist")

		var de errors.DetailedError
		require.Error(t, err)
		require.ErrorAs(t, err, &de)
		assert.True(t, de.IsNotFound())
	})

	t.Run("Create", func(t *testing.T) {
		createDataset := &client.Dataset{
			Name: datasetName,
		}
		d, err := c.Datasets.Create(ctx, createDataset)
		require.NoError(t, err)

		assert.NotNil(t, d.LastWrittenAt, "last written at is empty")
		assert.NotNil(t, d.CreatedAt, "created at is empty")
		// copy dynamic fields before asserting equality
		d.LastWrittenAt = currentDataset.LastWrittenAt
		d.CreatedAt = currentDataset.CreatedAt

		assert.Equal(t, currentDataset, d)
	})

	t.Run("Update", func(t *testing.T) {
		updatedDescription := "buzzing with data"
		updatedExpandJSONDepth := 3

		updateDataset := &client.Dataset{
			Name:            datasetName,
			Description:     updatedDescription,
			ExpandJSONDepth: updatedExpandJSONDepth,
		}
		t.Cleanup(func() {
			// revert updated fields to defaults after the test run
			//nolint:errcheck
			c.Datasets.Update(ctx, &client.Dataset{Name: datasetName})
		})
		d, err := c.Datasets.Update(ctx, updateDataset)
		require.NoError(t, err)
		assert.Equal(t, d.Description, updatedDescription)
		assert.Equal(t, d.ExpandJSONDepth, updatedExpandJSONDepth)
	})
}
