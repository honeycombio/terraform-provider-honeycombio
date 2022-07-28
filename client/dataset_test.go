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
	description := "new honeycomb dataset"
	expandJSONDepth := 0

	currentDataset := &Dataset{
		Name:            datasetName,
		Description:     &description,
		Slug:            urlEncodeDataset(datasetName),
		ExpandJSONDepth: &expandJSONDepth,
	}

	t.Run("List", func(t *testing.T) {
		d, err := c.Datasets.List(ctx)

		assert.NoError(t, err)
		assert.Contains(t, d, *currentDataset, "could not find current dataset with List")
	})

	t.Run("Get", func(t *testing.T) {
		d, err := c.Datasets.Get(ctx, currentDataset.Slug)

		assert.NoError(t, err)
		assert.Equal(t, *currentDataset, *d)
	})

	t.Run("Get_notFound", func(t *testing.T) {
		_, err := c.Datasets.Get(ctx, "does-not-exist")

		assert.Equal(t, ErrNotFound, err)
	})

	t.Run("Create", func(t *testing.T) {
		createDataset := &Dataset{
			Name:            datasetName,
			Description:     currentDataset.Description,
			ExpandJSONDepth: currentDataset.ExpandJSONDepth,
		}
		d, err := c.Datasets.Create(ctx, createDataset)

		assert.NoError(t, err)
		assert.Equal(t, currentDataset, d)
	})

	t.Run("Update", func(t *testing.T) {
		description := "brand new description"
		expandJSONDepth := 3
		updateDataset := &Dataset{
			Name:            datasetName,
			Description:     &description,
			ExpandJSONDepth: &expandJSONDepth,
		}
		d, err := c.Datasets.Update(ctx, updateDataset)

		assert.NoError(t, err)
		assert.Equal(t, currentDataset, d)
	})
}
