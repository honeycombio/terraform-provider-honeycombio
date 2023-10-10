package client

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

		require.NoError(t, err)

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

		var de DetailedError
		assert.Error(t, err)
		assert.ErrorAs(t, err, &de)
		assert.Equal(t, de.Status, http.StatusConflict)
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

	t.Run("Fail to Get Deleted DC", func(t *testing.T) {
		_, err := c.DerivedColumns.Get(ctx, dataset, derivedColumn.ID)

		var de DetailedError
		assert.Error(t, err)
		assert.ErrorAs(t, err, &de)
		assert.True(t, de.IsNotFound())

		_, err = c.DerivedColumns.GetByAlias(ctx, dataset, derivedColumn.Alias)
		assert.Error(t, err)
		assert.ErrorAs(t, err, &de)
		assert.True(t, de.IsNotFound())
	})
}
