package client_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
)

func TestQueryAnnotations(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	c := newTestClient(t)
	dataset := testDataset(t)

	var queryAnnotation *client.QueryAnnotation
	var err error

	// no cleanup func needed as queries cannot be deleted
	query, err := c.Queries.Create(ctx, dataset, &client.QuerySpec{
		Calculations: []client.CalculationSpec{
			{
				Op: "COUNT",
			},
		},
	})
	require.NoError(t, err)

	t.Run("Create", func(t *testing.T) {
		data := &client.QueryAnnotation{
			Name:        "Query created by a test",
			Description: "This derived column is created by a test",
			QueryID:     *query.ID,
		}
		queryAnnotation, err = c.QueryAnnotations.Create(ctx, dataset, data)

		require.NoError(t, err)
		data.ID = queryAnnotation.ID
		assert.Equal(t, data, queryAnnotation)
	})

	t.Run("List", func(t *testing.T) {
		queryAnnotations, err := c.QueryAnnotations.List(ctx, dataset)

		assert.NoError(t, err)
		assert.Contains(t, queryAnnotations, *queryAnnotation, "could not find QueryAnnotation with List")
	})

	t.Run("Get", func(t *testing.T) {
		result, err := c.QueryAnnotations.Get(ctx, dataset, queryAnnotation.ID)

		assert.NoError(t, err)
		assert.Equal(t, *queryAnnotation, *result)
	})

	t.Run("Update", func(t *testing.T) {
		// change all the fields to test
		data := &client.QueryAnnotation{
			ID:          queryAnnotation.ID,
			Name:        "This is a new name for the query created by a test",
			Description: "This is a new description",
			QueryID:     *query.ID,
		}
		queryAnnotation, err = c.QueryAnnotations.Update(ctx, dataset, data)

		assert.NoError(t, err)
		data.ID = queryAnnotation.ID
		assert.Equal(t, data, queryAnnotation)
	})

	t.Run("Delete", func(t *testing.T) {
		err = c.QueryAnnotations.Delete(ctx, dataset, queryAnnotation.ID)
		require.NoError(t, err)
	})

	t.Run("Fail to Get deleted Query Annotation", func(t *testing.T) {
		_, err := c.QueryAnnotations.Get(ctx, dataset, queryAnnotation.ID)

		var de client.DetailedError
		assert.Error(t, err)
		assert.ErrorAs(t, err, &de)
		assert.True(t, de.IsNotFound())
	})
}
