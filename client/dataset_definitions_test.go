package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDatasetDefinitions(t *testing.T) {
	ctx := context.Background()

	name := "dataset definition test"
	definition := "error" // sets the default behvior for error datatset definition
	definitionColumn := &DefinitionColumn{
		Name: &name,
		ID:   &definition,
	}

	datasetDefinition := &DatasetDefinition{
		Name: *definitionColumn,
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
		result, err := c.DatasetDefinitions.Create(ctx, dataset, *definitionColumn.Name, datasetDefinition)

		assert.NoError(t, err)
		assert.Equal(t, result, datasetDefinition)
	})

	t.Run("Update", func(t *testing.T) {
		updatedName := "actual error definition"
		updatedDefinition := "error"

		definitionColumn := &DefinitionColumn{
			Name: &updatedName,
			ID:   &updatedDefinition,
		}

		datasetDefinition := &DatasetDefinition{
			Name: *definitionColumn,
		}

		result, err := c.DatasetDefinitions.Update(ctx, dataset, *definitionColumn.Name, datasetDefinition)
		assert.NoError(t, err)
		assert.Equal(t, result, datasetDefinition)
	})

	t.Run("Delete", func(t *testing.T) {
		err = c.DerivedColumns.Delete(ctx, dataset, *datasetDefinition.Name.Name)

		assert.NoError(t, err)
	})
}
