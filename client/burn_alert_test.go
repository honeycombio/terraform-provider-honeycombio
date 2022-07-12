package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBurnAlerts(t *testing.T) {
	ctx := context.Background()

	var burnAlert *BurnAlert
	var err error

	c := newTestClient(t)
	dataset := testDataset(t)

	sli, err := c.DerivedColumns.Create(ctx, dataset, &DerivedColumn{
		Alias:      "sli.ba_slo_test",
		Expression: "LT($duration_ms, 1000)",
	})
	if err != nil {
		t.Fatal(err)
	}
	slo, err := c.SLOs.Create(ctx, dataset, &SLO{
		Name:             "BurnAlert Test SLO",
		TimePeriodDays:   7,
		TargetPerMillion: 999000,
		SLI:              SLIRef{Alias: sli.Alias},
	})
	if err != nil {
		t.Fatal(err)
	}
	// remove SLO and SLI at the end of the test run
	t.Cleanup(func() {
		c.SLOs.Delete(ctx, dataset, slo.ID)
		c.DerivedColumns.Delete(ctx, dataset, sli.ID)
	})

	t.Run("Create", func(t *testing.T) {
		data := &BurnAlert{
			ExhaustionMinutes: int(24 * 60), // 24 hours
			SLO:               SLORef{ID: slo.ID},
			Recipients: []NotificationRecipient{
				{
					Type:   "email",
					Target: "testalert@example.com",
				},
			},
		}

		burnAlert, err = c.BurnAlerts.Create(ctx, dataset, data)

		assert.NoError(t, err, "failed to create BurnAlert")
		assert.NotNil(t, burnAlert.ID, "BurnAlert ID is empty")
		assert.NotNil(t, burnAlert.CreatedAt, "created at is empty")
		assert.NotNil(t, burnAlert.UpdatedAt, "updated at is empty")
		// copy dynamic fields before asserting equality
		data.ID = burnAlert.ID
		data.CreatedAt = burnAlert.CreatedAt
		data.UpdatedAt = burnAlert.UpdatedAt
		data.Recipients[0].ID = burnAlert.Recipients[0].ID
		assert.Equal(t, data, burnAlert)
	})

	t.Run("Get", func(t *testing.T) {
		getBA, err := c.BurnAlerts.Get(ctx, dataset, burnAlert.ID)
		assert.NoError(t, err, "failed to get BurnAlert by ID")
		assert.Equal(t, burnAlert, getBA)
	})

	t.Run("Update", func(t *testing.T) {
		burnAlert.ExhaustionMinutes = int(4 * 60) // 4 hours

		result, err := c.BurnAlerts.Update(ctx, dataset, burnAlert)

		assert.NoError(t, err, "failed to update BurnAlert")
		// copy dynamic field before asserting equality
		burnAlert.UpdatedAt = result.UpdatedAt
		assert.Equal(t, burnAlert, result)
	})

	t.Run("ListForSLO", func(t *testing.T) {
		results, err := c.BurnAlerts.ListForSLO(ctx, dataset, slo.ID)
		burnAlert.Recipients = []NotificationRecipient{}
		assert.NoError(t, err, "failed to list burn alerts for SLO")
		assert.NotZero(t, len(results))
		assert.Equal(t, burnAlert.ID, results[0].ID, "newly created BurnAlert not in list of SLO's burn alerts")
	})

	t.Run("Delete", func(t *testing.T) {
		err = c.BurnAlerts.Delete(ctx, dataset, burnAlert.ID)

		assert.NoError(t, err, "failed to delete BurnAlert")
	})

	t.Run("Get_NotFound", func(t *testing.T) {
		_, err := c.BurnAlerts.Get(ctx, dataset, burnAlert.ID)

		assert.Equal(t, ErrNotFound, err)
	})
}
