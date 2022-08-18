package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDatasetDefinitions(t *testing.T) {
	ctx := context.Background()

	name := "dataset definition test"
	dcName := &DefinitionColumn{
		Name: &name,
	}

	datasetDefinition := &DatasetDefinition{
		Name: *dcName,
	}

	var err error

	c := newTestClient(t)
	dataset := testDataset(t)

	// Terraform Test Cases
	t.Run("List", func(t *testing.T) {
		result, err := c.DatasetDefinitions.List(ctx, dataset)

		assert.NoError(t, err)
		assert.Contains(t, result, *datasetDefinition, "could not find newly created definition with List")
	})

	t.Run("Get", func(t *testing.T) {
		result, err := c.DatasetDefinitions.Get(ctx, dataset, definition)

		assert.NoError(t, err)
		assert.Equal(t, *datasetDefinition, *result)
	})

	t.Run("Create", func(t *testing.T) {
		dd, err = c.DatasetDefinitions.Create(ctx, dataset, datasetDefinition)

		assert.NoError(t, err)

		data.ID = dd.ID
		assert.Equal(t, data, datasetDefinition)
	})

	t.Run("Update", func(t *testing.T) {
		datasetDefinition.Name = "new def name"

		// get DefinitionColumn.Name in [valid definitions]

		// extract, definition/value to be update for that definition

		result, err := c.DatasetDefinitions.Update(ctx, dataset, definition, value)

		assert.NoError(result, err)
	})

	t.Run("Delete", func(t *testing.T) {
		err = c.DerivedColumns.Delete(ctx, dataset, *datasetDefinition.Name.Name)

		assert.NoError(t, err)
	})
}
