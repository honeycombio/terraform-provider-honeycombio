package client_test

import (
	"context"
	"slices"
	"testing"
	"time"

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
	t.Cleanup(func() {
		// remove SLI DC at end of test run
		c.DerivedColumns.Delete(ctx, dataset, sli.ID)
	})

	t.Run("Create", func(t *testing.T) {
		data := &client.SLO{
			Name:             test.RandomStringWithPrefix("test.", 10),
			Description:      "My Super Sweet Test",
			TimePeriodDays:   30,
			TargetPerMillion: 995000,
			SLI:              client.SLIRef{Alias: sli.Alias},
			Tags: []client.Tag{
				{Key: "color", Value: "blue"},
				{Key: "team", Value: "a-team"},
			},
		}
		slo, err = c.SLOs.Create(ctx, dataset, data)
		require.NoError(t, err, "unable to create SLO")

		assert.NotNil(t, slo.ID, "SLO ID is empty")
		assert.Equal(t, slo.Name, data.Name)
		assert.Equal(t, slo.Description, data.Description)
		assert.Equal(t, slo.TimePeriodDays, data.TimePeriodDays)
		assert.Equal(t, slo.TargetPerMillion, data.TargetPerMillion)
		assert.Equal(t, slo.SLI.Alias, data.SLI.Alias)
		assert.WithinDuration(t, time.Now(), slo.CreatedAt, time.Second*15)
		assert.WithinDuration(t, time.Now(), slo.UpdatedAt, time.Second*15)
		assert.Equal(t, []string{dataset}, slo.DatasetSlugs)
		assert.ElementsMatch(t, slo.Tags, data.Tags, "tags do not match")
	})

	t.Run("List all slos for a dataset", func(t *testing.T) {
		// this has proven to be a bit racey after the create above, so we'll retry a few times
		assert.EventuallyWithT(t, func(col *assert.CollectT) {
			slos, err := c.SLOs.List(ctx, dataset)
			require.NoError(col, err)

			// not doing an Equal here because the timestamps may be different
			// and confirming the ID is in the listing is sufficient
			assert.Condition(col, func() bool {
				return slices.ContainsFunc(slos, func(s client.SLO) bool {
					return s.ID == slo.ID
				})
			})
		}, time.Second, 200*time.Millisecond, "could not find newly created SLO in List")
	})

	t.Run("List all slos in an environment", func(t *testing.T) {
		// this has proven to be a bit racey after the create above, so we'll retry a few times
		assert.EventuallyWithT(t, func(col *assert.CollectT) {
			slos, err := c.SLOs.List(ctx, client.EnvironmentWideSlug)
			require.NoError(col, err)

			// not doing an Equal here because the timestamps may be different
			// and confirming the ID is in the listing is sufficient
			assert.Condition(col, func() bool {
				return slices.ContainsFunc(slos, func(s client.SLO) bool {
					return s.ID == slo.ID
				})
			})
		}, time.Second, 200*time.Millisecond, "could not find newly created SLO in List")
	})

	t.Run("Get", func(t *testing.T) {
		result, err := c.SLOs.Get(ctx, dataset, slo.ID)
		require.NoError(t, err, "failed to get SLO by ID")
		assert.Equal(t, slo.ID, result.ID)
	})

	t.Run("Update", func(t *testing.T) {
		slo.Name = test.RandomStringWithPrefix("test.", 10)
		slo.TimePeriodDays = 14
		slo.Description = "Even sweeter"
		slo.TargetPerMillion = 990000
		slo.Tags = append(slo.Tags, client.Tag{Key: "new", Value: "tag"})

		result, err := c.SLOs.Update(ctx, dataset, slo)
		require.NoError(t, err, "failed to update SLO")

		assert.Equal(t, slo.ID, result.ID)
		assert.Equal(t, slo.Name, result.Name)
		assert.Equal(t, slo.Description, result.Description)
		assert.Equal(t, slo.TimePeriodDays, result.TimePeriodDays)
		assert.Equal(t, slo.TargetPerMillion, result.TargetPerMillion)
		assert.Equal(t, slo.SLI.Alias, result.SLI.Alias)
		assert.Equal(t, slo.CreatedAt, result.CreatedAt)
		assert.WithinDuration(t, time.Now(), slo.UpdatedAt, time.Second*15)
		assert.Equal(t, []string{dataset}, slo.DatasetSlugs)
		assert.ElementsMatch(t, slo.Tags, result.Tags, "tags do not match")
	})

	t.Run("Delete", func(t *testing.T) {
		err = c.SLOs.Delete(ctx, dataset, slo.ID)
		require.NoError(t, err, "failed to delete SLO")
	})

	t.Run("Fail to Get deleted SLO", func(t *testing.T) {
		_, err := c.SLOs.Get(ctx, dataset, slo.ID)

		var de client.DetailedError
		require.Error(t, err)
		require.ErrorAs(t, err, &de)
		assert.True(t, de.IsNotFound())
	})
}

func Test_MDSLOs(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	var err error

	c := newTestClient(t)

	if c.IsClassic(ctx) {
		t.Skip("Classic does not support multi-dataset SLOs")
	}

	// test setup for MD SLO
	var mdSLO *client.SLO
	var mdSLI *client.DerivedColumn
	var dataset1, dataset2 *client.Dataset

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

	mdSLI, err = c.DerivedColumns.Create(ctx, client.EnvironmentWideSlug, &client.DerivedColumn{
		Alias:       test.RandomStringWithPrefix("test.", 10),
		Description: "test SLI",
		Expression:  "BOOL(1)",
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		c.SLOs.Delete(ctx, client.EnvironmentWideSlug, mdSLO.ID)
		c.DerivedColumns.Delete(ctx, client.EnvironmentWideSlug, mdSLI.ID)

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

		mdSLO, err = c.SLOs.Create(ctx, client.EnvironmentWideSlug, mdData)

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

	t.Run("Get MD SLO", func(t *testing.T) {
		if c.IsClassic(ctx) {
			t.Skip("Classic does not support multi-dataset SLOs")
		}

		getMDSLO, err := c.SLOs.Get(ctx, client.EnvironmentWideSlug, mdSLO.ID)
		require.NoError(t, err, "failed to get MD SLO by ID")
		assert.Equal(t, *mdSLO, *getMDSLO)
	})

	t.Run("Update MD SLO", func(t *testing.T) {
		if c.IsClassic(ctx) {
			t.Skip("Classic does not support multi-dataset SLOs")
		}

		mdSLO.Name = test.RandomStringWithPrefix("test.", 10)
		mdSLO.TimePeriodDays = 14
		mdSLO.Description = "Even sweeter"
		mdSLO.TargetPerMillion = 990000

		result, err := c.SLOs.Update(ctx, client.EnvironmentWideSlug, mdSLO)

		require.NoError(t, err, "failed to update MD SLO")
		// copy dynamic field before asserting equality
		mdSLO.UpdatedAt = result.UpdatedAt
		assert.Equal(t, result, mdSLO)
	})

	t.Run("Delete MD SLO", func(t *testing.T) {
		err = c.SLOs.Delete(ctx, client.EnvironmentWideSlug, mdSLO.ID)
		require.NoError(t, err, "failed to delete MD SLO")
	})
}
