package client_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper"
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
		c.DerivedColumns.Delete(ctx, dataset, sli.ID)
	})

	// test setup for MD SLO
	var mdSLO *client.SLO
	var mdSLI *client.DerivedColumn
	var dataset1, dataset2 *client.Dataset
	if !c.IsClassic(ctx) {
		dataset1, err = c.Datasets.Create(ctx, &client.Dataset{
			Name:        test.RandomStringWithPrefix("test.", 10),
			Description: "test dataset 1",
		})
		require.NoError(t, err)

		dataset2, err = c.Datasets.Create(ctx, &client.Dataset{
			Name:        test.RandomStringWithPrefix("test.", 10),
			Description: "test dataset 2",
		})
		require.NoError(t, err)

		mdSLI, err = c.DerivedColumns.Create(ctx, client.Dataset_All, &client.DerivedColumn{
			Alias:       acctest.RandString(4) + "_sli",
			Description: "test SLI",
			Expression:  "BOOL(1)",
		})
		require.NoError(t, err)

		t.Cleanup(func() {
			c.SLOs.Delete(ctx, client.Dataset_All, mdSLO.ID)
			c.DerivedColumns.Delete(ctx, client.Dataset_All, sli.ID)

			c.Datasets.Update(ctx, &client.Dataset{
				Slug: dataset1.Slug,
				Settings: client.DatasetSettings{
					DeleteProtected: helper.ToPtr(false),
				},
			})
			err = c.Datasets.Delete(ctx, dataset1.Slug)

			c.Datasets.Update(ctx, &client.Dataset{
				Slug: dataset2.Slug,
				Settings: client.DatasetSettings{
					DeleteProtected: helper.ToPtr(false),
				},
			})
			err = c.Datasets.Delete(ctx, dataset2.Slug)
		})
	}

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
		data.DatasetSlugs = []string{dataset}

		assert.Equal(t, data, slo)
	})

	t.Run("Create MD SLO", func(t *testing.T) {
		if c.IsClassic(ctx) {
			t.Skip("Classic does not support multi-dataset SLOs")
		}

		mdData := &client.SLO{
			Name:             test.RandomStringWithPrefix("test.", 10),
			Description:      "My Super Sweet Test",
			TimePeriodDays:   30,
			TargetPerMillion: 995000,
			SLI:              client.SLIRef{Alias: mdSLI.Alias},
			DatasetSlugs:     []string{dataset1.Slug, dataset2.Slug},
		}

		mdSLO, err = c.SLOs.Create(ctx, client.Dataset_All, mdData)

		require.NoError(t, err, "unable to create SLO")
		assert.NotNil(t, mdSLO.ID, "SLO ID is empty")
		assert.NotNil(t, mdSLO.CreatedAt, "created at is empty")
		assert.NotNil(t, mdSLO.UpdatedAt, "updated at is empty")
		// copy dynamic fields before asserting equality
		mdData.ID = mdSLO.ID
		mdData.CreatedAt = mdSLO.CreatedAt
		mdData.UpdatedAt = mdSLO.UpdatedAt
		mdData.DatasetSlugs = []string{dataset1.Slug, dataset2.Slug}

		assert.Equal(t, mdData, mdSLO)
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

	t.Run("Get MD SLO", func(t *testing.T) {
		if c.IsClassic(ctx) {
			t.Skip("Classic does not support multi-dataset SLOs")
		}

		getMDSLO, err := c.SLOs.Get(ctx, client.Dataset_All, mdSLO.ID)
		require.NoError(t, err, "failed to get MD SLO by ID")
		assert.Equal(t, *mdSLO, *getMDSLO)
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

	t.Run(("Update MD SLO"), func(t *testing.T) {
		if c.IsClassic(ctx) {
			t.Skip("Classic does not support multi-dataset SLOs")
		}

		mdSLO.Name = test.RandomStringWithPrefix("test.", 10)
		mdSLO.TimePeriodDays = 14
		mdSLO.Description = "Even sweeter"
		mdSLO.TargetPerMillion = 990000

		result, err := c.SLOs.Update(ctx, client.Dataset_All, mdSLO)

		require.NoError(t, err, "failed to update MD SLO")
		// copy dynamic field before asserting equality
		mdSLO.UpdatedAt = result.UpdatedAt
		assert.Equal(t, result, mdSLO)
	})

	t.Run("Delete", func(t *testing.T) {
		err = c.SLOs.Delete(ctx, dataset, slo.ID)

		require.NoError(t, err, "failed to delete SLO")
	})

	t.Run("Delete MD SLO", func(t *testing.T) {
		if c.IsClassic(ctx) {
			t.Skip("Classic does not support multi-dataset SLOs")
		}

		err = c.SLOs.Delete(ctx, client.Dataset_All, mdSLO.ID)
		require.NoError(t, err, "failed to delete MD SLO")
	})

	t.Run("Fail to Get deleted SLO", func(t *testing.T) {
		_, err := c.SLOs.Get(ctx, dataset, slo.ID)

		var de client.DetailedError
		require.Error(t, err)
		require.ErrorAs(t, err, &de)
		assert.True(t, de.IsNotFound())
	})
}
