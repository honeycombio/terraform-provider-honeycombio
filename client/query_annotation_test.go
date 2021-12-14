package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueryAnnotations(t *testing.T) {
	ctx := context.Background()

	c := newTestClient(t)
	dataset := testDataset(t)

	var queryAnnotation *QueryAnnotation
	var err error

	query, err := c.Queries.Create(ctx, dataset, &QuerySpec{
		Calculations: []CalculationSpec{
			{
				Op: "COUNT",
			},
		},
	})

	if err != nil {
		t.Fatal(err)
	}

	t.Run("Create", func(t *testing.T) {
		data := &QueryAnnotation{
			Name:        "Query created by a test",
			Description: "This derived column is created by a test",
			QueryID:     *query.ID,
		}
		queryAnnotation, err = c.QueryAnnotations.Create(ctx, dataset, data)

		if err != nil {
			t.Fatal(err)
		}

		data.ID = queryAnnotation.ID
		assert.Equal(t, data, queryAnnotation)
	})

	t.Run("List", func(t *testing.T) {
		queryAnnotations, err := c.QueryAnnotations.List(ctx, dataset)

		if err != nil {
			t.Fatal(err)
		}

		assert.Contains(t, queryAnnotations, *queryAnnotation, "could not find QueryAnnotation with List")
	})

	t.Run("Get", func(t *testing.T) {
		result, err := c.QueryAnnotations.Get(ctx, dataset, queryAnnotation.ID)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, *queryAnnotation, *result)
	})

	t.Run("Update", func(t *testing.T) {
		// change all the fields to test
		data := &QueryAnnotation{
			ID:          queryAnnotation.ID,
			Name:        "This is a new name for the query created by a test",
			Description: "This is a new description",
			QueryID:     *query.ID,
		}
		queryAnnotation, err = c.QueryAnnotations.Update(ctx, dataset, data)

		if err != nil {
			t.Fatal(err)
		}

		data.ID = queryAnnotation.ID
		assert.Equal(t, data, queryAnnotation)
	})

	t.Run("Delete", func(t *testing.T) {
		err = c.QueryAnnotations.Delete(ctx, dataset, queryAnnotation.ID)

		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Get_notFound", func(t *testing.T) {
		_, err := c.QueryAnnotations.Get(ctx, dataset, queryAnnotation.ID)

		assert.Equal(t, ErrNotFound, err)
	})
}
