package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSLOs(t *testing.T) {
	ctx := context.Background()

	var slo *SLO
	var err error

	c := newTestClient(t)
	dataset := testDataset(t)

	sli, err := c.DerivedColumns.Create(ctx, dataset, &DerivedColumn{
		Alias:      "sli.slo_test",
		Expression: "LT($duration_ms, 1000)",
	})
	if err != nil {
		t.Fatal(err)
	}
	// remove SLI DC at end of test run
	t.Cleanup(func() {
		c.DerivedColumns.Delete(ctx, dataset, sli.ID)
	})

	t.Run("Create", func(t *testing.T) {
		data := &SLO{
			Name:             "Testsuite SLO",
			Description:      "My Super Sweet Test",
			TimePeriodDays:   30,
			TargetPerMillion: 995000,
			SLI:              SLIRef{Alias: sli.Alias},
		}

		slo, err = c.SLOs.Create(ctx, dataset, data)

		assert.NoError(t, err, "unable to create SLO")
		assert.NotNil(t, slo.ID, "SLO ID is empty")
		assert.NotNil(t, slo.CreatedAt, "created at is empty")
		assert.NotNil(t, slo.UpdatedAt, "updated at is empty")
		// copy dynamic fields before asserting equality
		data.ID = slo.ID
		data.CreatedAt = slo.CreatedAt
		data.UpdatedAt = slo.UpdatedAt
		assert.Equal(t, data, slo)
	})

	t.Run("List", func(t *testing.T) {
		results, err := c.SLOs.List(ctx, dataset)

		assert.NoError(t, err, "unable to list SLOs")
		assert.Contains(t, results, *slo, "could not find newly created SLO with List")
	})

	t.Run("Get", func(t *testing.T) {
		getSLO, err := c.SLOs.Get(ctx, dataset, slo.ID)

		assert.NoError(t, err, "failed to get SLO by ID")
		assert.Equal(t, *slo, *getSLO)
	})

	t.Run("Update", func(t *testing.T) {
		slo.Name = "Test Sweet SLO"
		slo.TimePeriodDays = 14
		slo.Description = "Even sweeter"
		slo.TargetPerMillion = 990000

		result, err := c.SLOs.Update(ctx, dataset, slo)

		assert.NoError(t, err, "failed to update SLO")
		// copy dynamic field before asserting equality
		slo.UpdatedAt = result.UpdatedAt
		assert.Equal(t, result, slo)
	})

	t.Run("Delete", func(t *testing.T) {
		err = c.SLOs.Delete(ctx, dataset, slo.ID)

		assert.NoError(t, err, "failed to delete SLO")
	})

	t.Run("Get_NotFound", func(t *testing.T) {
		_, err := c.SLOs.Get(ctx, dataset, slo.ID)

		assert.Equal(t, ErrNotFound, err)
	})
}
