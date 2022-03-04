package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDerivedColumns(t *testing.T) {
	ctx := context.Background()

	c := newTestClient(t)
	dataset := testDataset(t)

	var derivedColumn *DerivedColumn
	var err error

	t.Run("Create", func(t *testing.T) {
		data := &DerivedColumn{
			Alias:       "derived_column_test",
			Expression:  "LOG10($duration_ms)",
			Description: "This derived column is created by a test",
		}
		derivedColumn, err = c.DerivedColumns.Create(ctx, dataset, data)

		assert.NoError(t, err)

		data.ID = derivedColumn.ID
		assert.Equal(t, data, derivedColumn)
	})

	t.Run("Create_DuplicateErr", func(t *testing.T) {
		data := &DerivedColumn{
			Alias:       "derived_column_test",
			Expression:  "LOG10($duration_ms)",
			Description: "This is a derived column with the same name as an existing one",
		}
		_, err = c.DerivedColumns.Create(ctx, dataset, data)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "chosen alias is the same as an existing derived column")
	})

	t.Run("List", func(t *testing.T) {
		derivedColumns, err := c.DerivedColumns.List(ctx, dataset)

		assert.NoError(t, err)
		assert.Contains(t, derivedColumns, *derivedColumn, "could not find DerivedColumn with List")
	})

	t.Run("Get", func(t *testing.T) {
		result, err := c.DerivedColumns.Get(ctx, dataset, derivedColumn.ID)

		assert.NoError(t, err)
		assert.Equal(t, *derivedColumn, *result)
	})

	t.Run("GetByAlias", func(t *testing.T) {
		result, err := c.DerivedColumns.GetByAlias(ctx, dataset, derivedColumn.Alias)

		assert.NoError(t, err)
		assert.Equal(t, *derivedColumn, *result)
	})

	t.Run("Update", func(t *testing.T) {
		// change all the fields to test
		data := &DerivedColumn{
			ID:          derivedColumn.ID,
			Alias:       "derived_column_test_new_alias",
			Expression:  "DIV($duration_ms, 2)",
			Description: "This is a new description",
		}
		derivedColumn, err = c.DerivedColumns.Update(ctx, dataset, data)

		assert.NoError(t, err)

		data.ID = derivedColumn.ID
		assert.Equal(t, data, derivedColumn)
	})

	t.Run("Delete", func(t *testing.T) {
		err = c.DerivedColumns.Delete(ctx, dataset, derivedColumn.ID)

		assert.NoError(t, err)
	})

	t.Run("Get_notFound", func(t *testing.T) {
		_, err := c.DerivedColumns.Get(ctx, dataset, derivedColumn.ID)

		assert.Equal(t, ErrNotFound, err)
	})

	t.Run("GetByAlias_notFound", func(t *testing.T) {
		_, err := c.DerivedColumns.GetByAlias(ctx, dataset, derivedColumn.Alias)

		assert.Equal(t, ErrNotFound, err)
	})
}
