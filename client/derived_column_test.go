package client_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/test"
)

func TestDerivedColumns(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	c := newTestClient(t)
	dataset := testDataset(t)

	var derivedColumn *client.DerivedColumn
	var err error

	t.Run("Create", func(t *testing.T) {
		data := &client.DerivedColumn{
			Alias:       test.RandomStringWithPrefix("test.", 10),
			Expression:  "BOOL(1)",
			Description: "This derived column is created by a test",
		}
		derivedColumn, err = c.DerivedColumns.Create(ctx, dataset, data)

		require.NoError(t, err)

		data.ID = derivedColumn.ID
		assert.Equal(t, data, derivedColumn)
	})

	t.Run("Create_DuplicateErr", func(t *testing.T) {
		data := &client.DerivedColumn{
			Alias:       derivedColumn.Alias,
			Expression:  "BOOL(0)",
			Description: "This is a derived column with the same name as an existing one",
		}
		_, err = c.DerivedColumns.Create(ctx, dataset, data)

		var de client.DetailedError
		require.Error(t, err)
		require.ErrorAs(t, err, &de)
		assert.Equal(t, http.StatusConflict, de.Status)
	})

	t.Run("List", func(t *testing.T) {
		derivedColumns, err := c.DerivedColumns.List(ctx, dataset)

		require.NoError(t, err)
		assert.Contains(t, derivedColumns, *derivedColumn, "could not find DerivedColumn with List")
	})

	t.Run("Get", func(t *testing.T) {
		result, err := c.DerivedColumns.Get(ctx, dataset, derivedColumn.ID)

		require.NoError(t, err)
		assert.Equal(t, *derivedColumn, *result)
	})

	t.Run("GetByAlias", func(t *testing.T) {
		result, err := c.DerivedColumns.GetByAlias(ctx, dataset, derivedColumn.Alias)

		require.NoError(t, err)
		assert.Equal(t, *derivedColumn, *result)
	})

	t.Run("Update", func(t *testing.T) {
		// change all the fields to test
		data := &client.DerivedColumn{
			ID:          derivedColumn.ID,
			Alias:       test.RandomStringWithPrefix("test.", 10),
			Expression:  "BOOL(0)",
			Description: "This is a new description",
		}
		derivedColumn, err = c.DerivedColumns.Update(ctx, dataset, data)

		require.NoError(t, err)

		data.ID = derivedColumn.ID
		assert.Equal(t, data, derivedColumn)
	})

	t.Run("Delete", func(t *testing.T) {
		err = c.DerivedColumns.Delete(ctx, dataset, derivedColumn.ID)

		require.NoError(t, err)
	})

	t.Run("Fail to Get Deleted DC", func(t *testing.T) {
		_, err := c.DerivedColumns.Get(ctx, dataset, derivedColumn.ID)

		var de client.DetailedError
		require.Error(t, err)
		require.ErrorAs(t, err, &de)
		assert.True(t, de.IsNotFound())

		_, err = c.DerivedColumns.GetByAlias(ctx, dataset, derivedColumn.Alias)
		require.Error(t, err)
		require.ErrorAs(t, err, &de)
		assert.True(t, de.IsNotFound())
	})
}
