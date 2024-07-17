package client_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/test"
)

func TestBurnAlerts(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	var err error

	c := newTestClient(t)
	dataset := testDataset(t)
	testAlertEmail := test.RandomString(8) + "@example.com"

	sli, err := c.DerivedColumns.Create(ctx, dataset, &client.DerivedColumn{
		Alias:      test.RandomStringWithPrefix("test.", 8),
		Expression: "BOOL(1)",
	})
	require.NoError(t, err)
	slo, err := c.SLOs.Create(ctx, dataset, &client.SLO{
		Name:             test.RandomStringWithPrefix("test.", 8),
		TimePeriodDays:   7,
		TargetPerMillion: 999000,
		SLI:              client.SLIRef{Alias: sli.Alias},
	})
	require.NoError(t, err)

	// remove SLO and SLI at the end of the test run
	//nolint:errcheck
	t.Cleanup(func() {
		c.SLOs.Delete(ctx, dataset, slo.ID)
		c.DerivedColumns.Delete(ctx, dataset, sli.ID)

		// remove test alert email from recipients
		rcpts, _ := c.Recipients.List(ctx)
		for _, r := range rcpts {
			if r.Type == client.RecipientTypeEmail && r.Details.EmailAddress == testAlertEmail {
				c.Recipients.Delete(ctx, r.ID)
				break
			}
		}
	})

	var defaultBurnAlert *client.BurnAlert
	exhaustionMinutes24Hours := 24 * 60
	defaultBurnAlertCreateRequest := client.BurnAlert{
		ExhaustionMinutes: &exhaustionMinutes24Hours,
		SLO:               client.SLORef{ID: slo.ID},
		Recipients: []client.NotificationRecipient{
			{
				Type:   "email",
				Target: testAlertEmail,
			},
		},
	}
	exhaustionMinutes1Hour := 60
	defaultBurnAlertUpdateRequest := client.BurnAlert{
		ExhaustionMinutes: &exhaustionMinutes1Hour,
		SLO:               client.SLORef{ID: slo.ID},
		Recipients: []client.NotificationRecipient{
			{
				Type:   "email",
				Target: testAlertEmail,
			},
		},
	}

	var exhaustionTimeBurnAlert *client.BurnAlert
	exhaustionMinutes0Minutes := 0
	exhaustionTimeBurnAlertCreateRequest := client.BurnAlert{
		AlertType:         client.BurnAlertAlertTypeExhaustionTime,
		ExhaustionMinutes: &exhaustionMinutes0Minutes,
		SLO:               client.SLORef{ID: slo.ID},
		Recipients: []client.NotificationRecipient{
			{
				Type:   "email",
				Target: testAlertEmail,
			},
		},
	}
	exhaustionMinutes4Hours := 4 * 60
	exhaustionTimeBurnAlertUpdateRequest := client.BurnAlert{
		AlertType:         client.BurnAlertAlertTypeExhaustionTime,
		ExhaustionMinutes: &exhaustionMinutes4Hours,
		SLO:               client.SLORef{ID: slo.ID},
		Recipients: []client.NotificationRecipient{
			{
				Type:   "email",
				Target: testAlertEmail,
			},
		},
	}

	var budgetRateBurnAlert *client.BurnAlert
	budgetRateWindowMinutes1Hour := 60
	budgetRateDecreaseThresholdPerMillion1Percent := 10000
	budgetRateBurnAlertCreateRequest := client.BurnAlert{
		AlertType:                             client.BurnAlertAlertTypeBudgetRate,
		BudgetRateWindowMinutes:               &budgetRateWindowMinutes1Hour,
		BudgetRateDecreaseThresholdPerMillion: &budgetRateDecreaseThresholdPerMillion1Percent,
		SLO:                                   client.SLORef{ID: slo.ID},
		Recipients: []client.NotificationRecipient{
			{
				Type:   "email",
				Target: testAlertEmail,
			},
		},
	}
	budgetRateWindowMinutes2Hours := 2 * 60
	budgetRateDecreaseThresholdPerMillion5Percent := 10000
	budgetRateBurnAlertUpdateRequest := client.BurnAlert{
		AlertType:                             client.BurnAlertAlertTypeBudgetRate,
		BudgetRateWindowMinutes:               &budgetRateWindowMinutes2Hours,
		BudgetRateDecreaseThresholdPerMillion: &budgetRateDecreaseThresholdPerMillion5Percent,
		SLO:                                   client.SLORef{ID: slo.ID},
		Recipients: []client.NotificationRecipient{
			{
				Type:   "email",
				Target: testAlertEmail,
			},
		},
	}

	testCases := map[string]struct {
		alertType     client.BurnAlertAlertType
		createRequest client.BurnAlert
		updateRequest client.BurnAlert
		burnAlert     *client.BurnAlert
	}{
		"default - exhaustion_time": {
			alertType:     client.BurnAlertAlertTypeExhaustionTime,
			createRequest: defaultBurnAlertCreateRequest,
			updateRequest: defaultBurnAlertUpdateRequest,
			burnAlert:     defaultBurnAlert,
		},
		"exhaustion_time": {
			alertType:     client.BurnAlertAlertTypeExhaustionTime,
			createRequest: exhaustionTimeBurnAlertCreateRequest,
			updateRequest: exhaustionTimeBurnAlertUpdateRequest,
			burnAlert:     exhaustionTimeBurnAlert,
		},
		"budget_rate": {
			alertType:     client.BurnAlertAlertTypeBudgetRate,
			createRequest: budgetRateBurnAlertCreateRequest,
			updateRequest: budgetRateBurnAlertUpdateRequest,
			burnAlert:     budgetRateBurnAlert,
		},
	}

	for testName, testCase := range testCases {
		var burnAlert *client.BurnAlert
		var err error

		t.Run(fmt.Sprintf("Create: %s", testName), func(t *testing.T) {
			data := &testCase.createRequest
			burnAlert, err = c.BurnAlerts.Create(ctx, dataset, data)
			require.NoError(t, err, "failed to create BurnAlert")
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
			require.NoError(t, err, "failed to get BurnAlert by ID")
			assert.Equal(t, burnAlert, result)
		})

		t.Run(fmt.Sprintf("Update: %s", testName), func(t *testing.T) {
			data := &testCase.updateRequest
			data.ID = burnAlert.ID

			burnAlert, err = c.BurnAlerts.Update(ctx, dataset, data)
			require.NoError(t, err, "failed to update BurnAlert")

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
			require.NoError(t, err, "failed to list burn alerts for SLO")

			assert.NotZero(t, len(results))
			assert.Equal(t, burnAlert.ID, results[0].ID, "newly created BurnAlert not in list of SLO's burn alerts")
		})

		t.Run(fmt.Sprintf("Delete - %s", testName), func(t *testing.T) {
			err = c.BurnAlerts.Delete(ctx, dataset, burnAlert.ID)

			require.NoError(t, err, "failed to delete BurnAlert")
		})

		t.Run(fmt.Sprintf("Fail to GET a deleted burn alert: %s", testName), func(t *testing.T) {
			_, err := c.BurnAlerts.Get(ctx, dataset, burnAlert.ID)

			var de client.DetailedError
			require.Error(t, err)
			require.ErrorAs(t, err, &de)
			assert.True(t, de.IsNotFound())
		})
	}
}

func TestBurnAlerts_BurnAlertAlertTypes(t *testing.T) {
	expectedAlertTypes := []client.BurnAlertAlertType{
		client.BurnAlertAlertTypeExhaustionTime,
		client.BurnAlertAlertTypeBudgetRate,
	}

	t.Run("returns expected burn alert alert types", func(t *testing.T) {
		actualAlertTypes := client.BurnAlertAlertTypes()

		assert.NotEmpty(t, actualAlertTypes)
		assert.Equal(t, len(expectedAlertTypes), len(actualAlertTypes))
		assert.ElementsMatch(t, expectedAlertTypes, actualAlertTypes)
	})
}
