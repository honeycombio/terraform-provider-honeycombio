package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDatasets(t *testing.T) {
	ctx := context.Background()

	c := newTestClient(t)
	datasetName := testDataset(t)

	currentDataset := &Dataset{
		Name: datasetName,
		Slug: urlEncodeDataset(datasetName),
	}

	// create a new dataset with the parameters above
	t.Run("Create", func(t *testing.T) {
		d, err := c.Datasets.Create(ctx, currentDataset)
		assert.NoError(t, err)
		assert.NotNil(t, d.LastWrittenAt, "last written at is empty")
		assert.NotNil(t, d.CreatedAt, "created at is empty")
		// copy dynamic fields before asserting - will be skipped if expected dataset not found
		if d.Name == currentDataset.Name {
			currentDataset.LastWrittenAt = d.LastWrittenAt
			currentDataset.CreatedAt = d.CreatedAt
		}
		assert.Equal(t, *currentDataset, *d)
	})

	t.Run("List", func(t *testing.T) {
		d, err := c.Datasets.List(ctx)

		assert.NoError(t, err)

		for _, dataset := range d {
			assert.NotNil(t, dataset.LastWrittenAt, "last written at is empty")
			assert.NotNil(t, dataset.CreatedAt, "created at is empty")
			// copy dynamic fields before asserting - will be skipped if expected dataset not found
			if dataset.Name == currentDataset.Name {
				currentDataset.LastWrittenAt = dataset.LastWrittenAt
				currentDataset.CreatedAt = dataset.CreatedAt
			}
		}
		assert.Contains(t, d, *currentDataset, "could not find current dataset with List")
	})

	t.Run("Get", func(t *testing.T) {
		d, err := c.Datasets.Get(ctx, currentDataset.Slug)

		assert.NoError(t, err)

		assert.NotNil(t, d.LastWrittenAt, "last written at is empty")
		assert.NotNil(t, d.CreatedAt, "created at is empty")
		// copy dynamic fields before asserting equality
		d.LastWrittenAt = currentDataset.LastWrittenAt
		d.CreatedAt = currentDataset.CreatedAt

		assert.Equal(t, *currentDataset, *d)
	})

	t.Run("Get_notFound", func(t *testing.T) {
		_, err := c.Datasets.Get(ctx, "does-not-exist")

		assert.Equal(t, ErrNotFound, err)
	})

	t.Run("Create", func(t *testing.T) {
		createDataset := &Dataset{
			Name: datasetName,
		}
		d, err := c.Datasets.Create(ctx, createDataset)

		assert.NoError(t, err)

		assert.NotNil(t, d.LastWrittenAt, "last written at is empty")
		assert.NotNil(t, d.CreatedAt, "created at is empty")
		// copy dynamic fields before asserting equality
		d.LastWrittenAt = currentDataset.LastWrittenAt
		d.CreatedAt = currentDataset.CreatedAt

		assert.Equal(t, currentDataset, d)
	})
}
