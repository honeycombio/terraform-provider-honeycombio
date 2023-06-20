package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestColumns(t *testing.T) {
	ctx := context.Background()

	c := newTestClient(t)
	dataset := testDataset(t)

	var column *Column
	var err error

	t.Run("Create", func(t *testing.T) {
		data := &Column{
			KeyName:     "column_test",
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
		result, err := c.Columns.GetByKeyName(ctx, dataset, "column_test")

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

	t.Run("Get_notFound", func(t *testing.T) {
		_, err := c.Columns.Get(ctx, dataset, column.ID)

		assert.Equal(t, ErrNotFound, err)
	})
}
