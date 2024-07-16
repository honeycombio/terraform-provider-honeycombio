package client_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/client/errors"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/test"
)

func TestSLOs(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	var slo *client.SLO
	var err error

	c := newTestClient(t)
	dataset := testDataset(t)

	sli, err := c.DerivedColumns.Create(ctx, dataset, &client.DerivedColumn{
		Alias:      test.RandomStringWithPrefix("test.", 10),
		Expression: "BOOL(1)",
	})
	require.NoError(t, err, "unable to create SLI")

	// remove SLI DC at end of test run
	t.Cleanup(func() {
		//nolint:errcheck
		c.DerivedColumns.Delete(ctx, dataset, sli.ID)
	})

	t.Run("Create", func(t *testing.T) {
		data := &client.SLO{
			Name:             test.RandomStringWithPrefix("test.", 10),
			Description:      "My Super Sweet Test",
			TimePeriodDays:   30,
			TargetPerMillion: 995000,
			SLI:              client.SLIRef{Alias: sli.Alias},
		}
		slo, err = c.SLOs.Create(ctx, dataset, data)

		require.NoError(t, err, "unable to create SLO")
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

		require.NoError(t, err, "unable to list SLOs")
		assert.Contains(t, results, *slo, "could not find newly created SLO with List")
	})

	t.Run("Get", func(t *testing.T) {
		getSLO, err := c.SLOs.Get(ctx, dataset, slo.ID)

		require.NoError(t, err, "failed to get SLO by ID")
		assert.Equal(t, *slo, *getSLO)
	})

	t.Run("Update", func(t *testing.T) {
		slo.Name = test.RandomStringWithPrefix("test.", 10)
		slo.TimePeriodDays = 14
		slo.Description = "Even sweeter"
		slo.TargetPerMillion = 990000

		result, err := c.SLOs.Update(ctx, dataset, slo)

		require.NoError(t, err, "failed to update SLO")
		// copy dynamic field before asserting equality
		slo.UpdatedAt = result.UpdatedAt
		assert.Equal(t, result, slo)
	})

	t.Run("Delete", func(t *testing.T) {
		err = c.SLOs.Delete(ctx, dataset, slo.ID)

		require.NoError(t, err, "failed to delete SLO")
	})

	t.Run("Fail to Get deleted SLO", func(t *testing.T) {
		_, err := c.SLOs.Get(ctx, dataset, slo.ID)

		var de errors.DetailedError
		require.Error(t, err)
		require.ErrorAs(t, err, &de)
		assert.True(t, de.IsNotFound())
	})
}
