package client

import (
	"context"
	"testing"

	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestColumns(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	c := newTestClient(t)
	dataset := testDataset(t)

	var column *Column
	var err error

	t.Run("Create", func(t *testing.T) {
		data := &Column{
			KeyName:     test.RandomStringWithPrefix("test.", 10),
			Hidden:      ToPtr(false),
			Description: "This column is created by a test",
			Type:        ToPtr(ColumnTypeFloat),
		}
		column, err = c.Columns.Create(ctx, dataset, data)

		assert.NoError(t, err)

		data.ID = column.ID
		assert.NotNil(t, column.LastWrittenAt, "last written at is empty")
		assert.NotNil(t, column.CreatedAt, "created at is empty")
		assert.NotNil(t, column.UpdatedAt, "updated at is empty")
		// copy dynamic fields before asserting equality
		data.LastWrittenAt = column.LastWrittenAt
		data.CreatedAt = column.CreatedAt
		data.UpdatedAt = column.UpdatedAt
		assert.Equal(t, data, column)
	})

	t.Run("List", func(t *testing.T) {
		columns, err := c.Columns.List(ctx, dataset)

		assert.NoError(t, err)
		assert.Contains(t, columns, *column, "could not find column with List")
	})

	t.Run("Get", func(t *testing.T) {
		result, err := c.Columns.Get(ctx, dataset, column.ID)

		assert.NoError(t, err)
		assert.Equal(t, *column, *result)
	})

	t.Run("GetByKeyName", func(t *testing.T) {
		result, err := c.Columns.GetByKeyName(ctx, dataset, column.KeyName)

		assert.NoError(t, err)
		assert.Equal(t, *column, *result)
	})

	t.Run("Update", func(t *testing.T) {
		// change all the fields to test
		data := &Column{
			ID:          column.ID,
			KeyName:     column.KeyName,
			Hidden:      ToPtr(true),
			Description: "This is a new description",
			Type:        ToPtr(ColumnTypeBoolean),
		}
		column, err = c.Columns.Update(ctx, dataset, data)

		assert.NoError(t, err)

		data.ID = column.ID
		assert.Equal(t, column.Description, data.Description)
		assert.Equal(t, column.Type, data.Type)
		assert.True(t, *column.Hidden)
		assert.NotNil(t, column.LastWrittenAt, "last written at is empty")
		assert.NotNil(t, column.CreatedAt, "created at is empty")
		assert.NotNil(t, column.UpdatedAt, "updated at is empty")
		// copy dynamic fields before asserting equality
		data.LastWrittenAt = column.LastWrittenAt
		data.CreatedAt = column.CreatedAt
		data.UpdatedAt = column.UpdatedAt
	})

	t.Run("Delete", func(t *testing.T) {
		err = c.Columns.Delete(ctx, dataset, column.ID)

		assert.NoError(t, err)
	})

	t.Run("Fail to get deleted Column", func(t *testing.T) {
		_, err := c.Columns.Get(ctx, dataset, column.ID)

		var de DetailedError
		assert.Error(t, err)
		assert.ErrorAs(t, err, &de)
		assert.True(t, de.IsNotFound())
	})
}

// createRandomTestColumns creates three columns with random names and returns them.
// One column is of type Float, and two are of type String.
//
// The columns are automatically cleaned up after the test run.
func createRandomTestColumns(t *testing.T, c *Client, dataset string) (*Column, *Column, *Column) {
	t.Helper()

	ctx := context.Background()

	floatCol, err := c.Columns.Create(ctx, dataset, &Column{
		KeyName: test.RandomStringWithPrefix("test.", 8),
		Type:    ToPtr(ColumnTypeFloat),
	})
	require.NoError(t, err)
	col1, err := c.Columns.Create(ctx, dataset, &Column{
		KeyName: test.RandomStringWithPrefix("test.", 8),
		Type:    ToPtr(ColumnTypeString),
	})
	require.NoError(t, err)
	col2, err := c.Columns.Create(ctx, dataset, &Column{
		KeyName: test.RandomStringWithPrefix("test.", 8),
		Type:    ToPtr(ColumnTypeString),
	})
	require.NoError(t, err)

	//nolint:errcheck
	t.Cleanup(func() {
		c.Columns.Delete(ctx, dataset, floatCol.ID)
		c.Columns.Delete(ctx, dataset, col1.ID)
		c.Columns.Delete(ctx, dataset, col2.ID)
	})

	return floatCol, col1, col2
}
