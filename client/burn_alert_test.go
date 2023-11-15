package client

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBurnAlerts(t *testing.T) {
	ctx := context.Background()

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
	//nolint:errcheck
	t.Cleanup(func() {
		c.SLOs.Delete(ctx, dataset, slo.ID)
		c.DerivedColumns.Delete(ctx, dataset, sli.ID)
	})

	var defaultBurnAlert *BurnAlert
	exhaustionMinutes24Hours := 24 * 60
	defaultBurnAlertCreateRequest := BurnAlert{
		ExhaustionMinutes: &exhaustionMinutes24Hours,
		SLO:               SLORef{ID: slo.ID},
		Recipients: []NotificationRecipient{
			{
				Type:   "email",
				Target: "testalert@example.com",
			},
		},
	}
	exhaustionMinutes1Hour := 60
	defaultBurnAlertUpdateRequest := BurnAlert{
		ExhaustionMinutes: &exhaustionMinutes1Hour,
		SLO:               SLORef{ID: slo.ID},
		Recipients: []NotificationRecipient{
			{
				Type:   "email",
				Target: "testalert@example.com",
			},
		},
	}

	var exhaustionTimeBurnAlert *BurnAlert
	exhaustionMinutes0Minutes := 0
	exhaustionTimeBurnAlertCreateRequest := BurnAlert{
		AlertType:         string(BurnAlertAlertTypeExhaustionTime),
		ExhaustionMinutes: &exhaustionMinutes0Minutes,
		SLO:               SLORef{ID: slo.ID},
		Recipients: []NotificationRecipient{
			{
				Type:   "email",
				Target: "testalert@example.com",
			},
		},
	}
	exhaustionMinutes4Hours := 4 * 60
	exhaustionTimeBurnAlertUpdateRequest := BurnAlert{
		AlertType:         string(BurnAlertAlertTypeExhaustionTime),
		ExhaustionMinutes: &exhaustionMinutes4Hours,
		SLO:               SLORef{ID: slo.ID},
		Recipients: []NotificationRecipient{
			{
				Type:   "email",
				Target: "testalert@example.com",
			},
		},
	}

	var budgetRateBurnAlert *BurnAlert
	budgetRateWindowMinutes1Hour := 60
	budgetRateDecreaseThresholdPerMillion1Percent := 10000
	budgetRateBurnAlertCreateRequest := BurnAlert{
		AlertType:                             string(BurnAlertAlertTypeBudgetRate),
		BudgetRateWindowMinutes:               &budgetRateWindowMinutes1Hour,
		BudgetRateDecreaseThresholdPerMillion: &budgetRateDecreaseThresholdPerMillion1Percent,
		SLO:                                   SLORef{ID: slo.ID},
		Recipients: []NotificationRecipient{
			{
				Type:   "email",
				Target: "testalert@example.com",
			},
		},
	}
	budgetRateWindowMinutes2Hours := 2 * 60
	budgetRateDecreaseThresholdPerMillion5Percent := 10000
	budgetRateBurnAlertUpdateRequest := BurnAlert{
		AlertType:                             string(BurnAlertAlertTypeBudgetRate),
		BudgetRateWindowMinutes:               &budgetRateWindowMinutes2Hours,
		BudgetRateDecreaseThresholdPerMillion: &budgetRateDecreaseThresholdPerMillion5Percent,
		SLO:                                   SLORef{ID: slo.ID},
		Recipients: []NotificationRecipient{
			{
				Type:   "email",
				Target: "testalert@example.com",
			},
		},
	}

	testCases := map[string]struct {
		alertType     string
		createRequest BurnAlert
		updateRequest BurnAlert
		burnAlert     *BurnAlert
	}{
		"default - exhaustion_time": {
			alertType:     string(BurnAlertAlertTypeExhaustionTime),
			createRequest: defaultBurnAlertCreateRequest,
			updateRequest: defaultBurnAlertUpdateRequest,
			burnAlert:     defaultBurnAlert,
		},
		"exhaustion_time": {
			alertType:     string(BurnAlertAlertTypeExhaustionTime),
			createRequest: exhaustionTimeBurnAlertCreateRequest,
			updateRequest: exhaustionTimeBurnAlertUpdateRequest,
			burnAlert:     exhaustionTimeBurnAlert,
		},
		"budget_rate": {
			alertType:     string(BurnAlertAlertTypeBudgetRate),
			createRequest: budgetRateBurnAlertCreateRequest,
			updateRequest: budgetRateBurnAlertUpdateRequest,
			burnAlert:     budgetRateBurnAlert,
		},
	}

	for testName, testCase := range testCases {
		var burnAlert *BurnAlert
		var err error

		t.Run(fmt.Sprintf("Create: %s", testName), func(t *testing.T) {
			data := &testCase.createRequest
			burnAlert, err = c.BurnAlerts.Create(ctx, dataset, data)

			assert.NoError(t, err, "failed to create BurnAlert")
			assert.NotNil(t, burnAlert.ID, "BurnAlert ID is empty")
			assert.NotNil(t, burnAlert.CreatedAt, "created at is empty")
			assert.NotNil(t, burnAlert.UpdatedAt, "updated at is empty")
			assert.Equal(t, testCase.alertType, burnAlert.AlertType)

			// copy dynamic fields before asserting equality
			data.AlertType = burnAlert.AlertType
			data.ID = burnAlert.ID
			data.CreatedAt = burnAlert.CreatedAt
			data.UpdatedAt = burnAlert.UpdatedAt
			data.Recipients[0].ID = burnAlert.Recipients[0].ID
			assert.Equal(t, data, burnAlert)
		})

		t.Run(fmt.Sprintf("Get: %s", testName), func(t *testing.T) {
			result, err := c.BurnAlerts.Get(ctx, dataset, burnAlert.ID)
			assert.NoError(t, err, "failed to get BurnAlert by ID")
			assert.Equal(t, burnAlert, result)
		})

		t.Run(fmt.Sprintf("Update: %s", testName), func(t *testing.T) {
			data := &testCase.updateRequest
			data.ID = burnAlert.ID

			burnAlert, err = c.BurnAlerts.Update(ctx, dataset, data)

			assert.NoError(t, err, "failed to update BurnAlert")

			// copy dynamic field before asserting equality
			data.AlertType = burnAlert.AlertType
			data.ID = burnAlert.ID
			data.CreatedAt = burnAlert.CreatedAt
			data.UpdatedAt = burnAlert.UpdatedAt
			data.Recipients[0].ID = burnAlert.Recipients[0].ID
			assert.Equal(t, burnAlert, data)
		})

		t.Run(fmt.Sprintf("ListForSLO: %s", testName), func(t *testing.T) {
			results, err := c.BurnAlerts.ListForSLO(ctx, dataset, slo.ID)

			assert.NoError(t, err, "failed to list burn alerts for SLO")
			assert.NotZero(t, len(results))
			assert.Equal(t, burnAlert.ID, results[0].ID, "newly created BurnAlert not in list of SLO's burn alerts")
		})

		t.Run(fmt.Sprintf("Delete - %s", testName), func(t *testing.T) {
			err = c.BurnAlerts.Delete(ctx, dataset, burnAlert.ID)

			assert.NoError(t, err, "failed to delete BurnAlert")
		})

		t.Run(fmt.Sprintf("Fail to GET a deleted burn alert: %s", testName), func(t *testing.T) {
			_, err := c.BurnAlerts.Get(ctx, dataset, burnAlert.ID)

			var de DetailedError
			assert.Error(t, err)
			assert.ErrorAs(t, err, &de)
			assert.True(t, de.IsNotFound())
		})
	}
}

func TestBurnAlerts_BurnAlertAlertTypes(t *testing.T) {
	expectedAlertTypes := []BurnAlertAlertType{
		BurnAlertAlertTypeExhaustionTime,
		BurnAlertAlertTypeBudgetRate,
	}

	t.Run("returns expected burn alert alert types", func(t *testing.T) {
		actualAlertTypes := BurnAlertAlertTypes()

		assert.NotEmpty(t, actualAlertTypes)
		assert.Equal(t, len(expectedAlertTypes), len(actualAlertTypes))
		assert.ElementsMatch(t, expectedAlertTypes, actualAlertTypes)
	})
}
