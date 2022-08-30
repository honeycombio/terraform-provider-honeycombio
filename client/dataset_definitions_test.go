package client

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// update, validate and clear a dataset definition
func TestDatasetDefinitions(t *testing.T) {
	ctx := context.Background()

	definitionName := "trace_id"
	//defintionValue := "datasetDef1"

	// Get the name/type of the Dataset Definition
	datasetDefinitionColumn := &DefinitionColumn{
		Name: &definitionName,
	}

	// Empty Definition to popluate after validation later
	datasetDefinition := &DatasetDefinition{}

	// validate it is an allowed field
	if ValidateDatasetDefinition(*datasetDefinitionColumn.Name) {
		// extract the Name Value and convert into proper DatasetDefintion value
		// for now hardcode TraceID - as this is the value we will initiall test
		datasetDefinition.TraceID = datasetDefinitionColumn
	} else {
		fmt.Printf("test - definition is invalid")
	}

	c := newTestClient(t)
	dataset := testDataset(t)

	t.Run("List", func(t *testing.T) {
		result, err := c.DatasetDefinitions.List(ctx, dataset)
		assert.NoError(t, err)

		for _, v := range result {
			assert.Contains(t, v, *datasetDefinition, "could not find newly created definition with List")
		}
	})

	t.Run("Update", func(t *testing.T) {
		updatedName := "actual error definition"
		//updatedDefinition := "error"

		definitionColumn := &DefinitionColumn{
			Name: &updatedName,
		}

		// hardcode trace ID for now
		datasetDefinition := &DatasetDefinition{
			TraceID: definitionColumn,
		}

		result, err := c.DatasetDefinitions.Update(ctx, dataset, datasetDefinition)
		assert.NoError(t, err)
		assert.Equal(t, result, datasetDefinition)
		assert.Equal(t, result.TraceID, "trace_id")
	})
}
