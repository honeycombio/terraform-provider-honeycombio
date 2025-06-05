package client_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/test"
)

func TestTriggers(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	var trigger *client.Trigger
	var err error

	c := newTestClient(t)
	dataset := testDataset(t)
	floatCol, col1, col2 := createRandomTestColumns(t, c, dataset)

	t.Run("Create", func(t *testing.T) {
		data := &client.Trigger{
			Name:        test.RandomStringWithPrefix("test.", 8),
			Description: "Some description",
			Disabled:    true,
			Query: &client.QuerySpec{
				Breakdowns: nil,
				Calculations: []client.CalculationSpec{
					{
						Op:     client.CalculationOpP99,
						Column: &floatCol.KeyName,
					},
				},
				Filters: []client.FilterSpec{
					{
						Column: col1.KeyName,
						Op:     client.FilterOpExists,
					},
					{
						Column: col2.KeyName,
						Op:     client.FilterOpContains,
						Value:  "foobar",
					},
				},
				FilterCombination: client.FilterCombinationOr,
			},
			Frequency: 300,
			Threshold: &client.TriggerThreshold{
				Op:    client.TriggerThresholdOpGreaterThan,
				Value: 10000,
			},
			Recipients: []client.NotificationRecipient{
				{
					Type:   client.RecipientTypeMarker,
					Target: "This marker is created by a trigger",
				},
			},
			Tags: []client.Tag{
				{Key: "color", Value: "blue"},
			},
		}
		trigger, err = c.Triggers.Create(ctx, dataset, data)
		require.NoError(t, err)
		assert.NotNil(t, trigger.ID)

		// copy IDs before asserting equality
		data.ID = trigger.ID
		data.QueryID = trigger.QueryID
		data.AlertType = trigger.AlertType
		data.Recipients[0].ID = trigger.Recipients[0].ID
		// set default time range
		data.Query.TimeRange = client.ToPtr(300)

		// set the default alert type
		data.AlertType = client.TriggerAlertTypeOnChange
		// set the default evaluation window type
		data.EvaluationScheduleType = client.TriggerEvaluationScheduleFrequency
		// set the default threshold exceeded limit
		data.Threshold.ExceededLimit = 1

		assert.NotEmpty(t, trigger.Tags)
		assert.ElementsMatch(t, trigger.Tags, data.Tags, "tags do not match")
		// trigger.Tags = data.Tags

		assert.Equal(t, data, trigger)
	})

	t.Run("List", func(t *testing.T) {
		result, err := c.Triggers.List(ctx, dataset)

		require.NoError(t, err)
		assert.Contains(t, result, *trigger, "could not find newly created trigger with List")
	})

	t.Run("Get", func(t *testing.T) {
		getTrigger, err := c.Triggers.Get(ctx, dataset, trigger.ID)

		require.NoError(t, err)
		assert.Equal(t, *trigger, *getTrigger)
	})

	t.Run("Update", func(t *testing.T) {
		trigger.Description = "A new description"

		// update the alert type to on true
		trigger.AlertType = client.TriggerAlertTypeOnTrue
		// update the evaluation schedule to a window type
		trigger.EvaluationScheduleType = client.TriggerEvaluationScheduleWindow
		trigger.EvaluationSchedule = &client.TriggerEvaluationSchedule{
			Window: client.TriggerEvaluationWindow{
				DaysOfWeek: []string{"monday", "wednesday", "friday"},
				StartTime:  "13:00",
				EndTime:    "21:00",
			},
		}
		trigger.Tags = []client.Tag{
			{Key: "team", Value: "t-team"},
		}
		// update the threshold exceeded limit to 3
		trigger.Threshold.ExceededLimit = 3

		result, err := c.Triggers.Update(ctx, dataset, trigger)

		// copy IDs before asserting equality
		trigger.QueryID = result.QueryID
		require.NoError(t, err)
		require.ElementsMatch(t, trigger.Tags, result.Tags, "tags do not match")
		assert.Equal(t, trigger, result, "full trigger does not match")
	})

	t.Run("Delete", func(t *testing.T) {
		err = c.Triggers.Delete(ctx, dataset, trigger.ID)

		require.NoError(t, err)
	})

	t.Run("Fail to Get deleted Trigger", func(t *testing.T) {
		_, err := c.Triggers.Get(ctx, dataset, trigger.ID)

		var de client.DetailedError
		require.Error(t, err)
		require.ErrorAs(t, err, &de)
		assert.True(t, de.IsNotFound())
	})
}

func TestMatchesTriggerSubset(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in          client.QuerySpec
		expectedErr error
	}{
		{
			in: client.QuerySpec{
				Calculations: []client.CalculationSpec{
					{
						Op: client.CalculationOpCount,
					},
				},
			},
			expectedErr: nil,
		},
		{
			in: client.QuerySpec{
				Calculations: nil,
			},
			expectedErr: errors.New("a trigger query should contain exactly one calculation"),
		},
		{
			in: client.QuerySpec{
				Calculations: []client.CalculationSpec{
					{
						Op: client.CalculationOpHeatmap,
					},
				},
			},
			expectedErr: errors.New("a trigger query may not contain a HEATMAP calculation"),
		},
		{
			in: client.QuerySpec{
				Calculations: []client.CalculationSpec{
					{
						Op: client.CalculationOpCount,
					},
				},
				Limit: client.ToPtr(100),
			},
			expectedErr: errors.New("limit is not allowed in a trigger query"),
		},
		{
			in: client.QuerySpec{
				Calculations: []client.CalculationSpec{
					{
						Op: client.CalculationOpCount,
					},
				},
				Orders: []client.OrderSpec{
					{
						Column: client.ToPtr(test.RandomStringWithPrefix("test.", 8)),
					},
				},
			},
			expectedErr: errors.New("orders is not allowed in a trigger query"),
		},
	}

	for i, c := range cases {
		err := client.MatchesTriggerSubset(&c.in)

		assert.Equal(t, c.expectedErr, err, "Test case %d, QuerySpec: %v", i, c.in)
	}
}

func TestTriggersWithBaselineDetails(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	var trigger *client.Trigger
	var err error

	c := newTestClient(t)
	dataset := testDataset(t)
	floatCol, _, _ := createRandomTestColumns(t, c, dataset)

	// delete the trigger after the tests run
	t.Cleanup(func() {
		c.Triggers.Delete(ctx, dataset, trigger.ID)
	})

	t.Run("Create - with baseline_details", func(t *testing.T) {
		data := &client.Trigger{
			Name:        test.RandomStringWithPrefix("test.", 8),
			Description: "Some description",
			Disabled:    true,
			Query: &client.QuerySpec{
				Breakdowns: nil,
				Calculations: []client.CalculationSpec{
					{
						Op:     client.CalculationOpRateAvg,
						Column: &floatCol.KeyName,
					},
				},
				TimeRange: client.ToPtr(900),
			},
			Frequency: 900,
			Threshold: &client.TriggerThreshold{
				Op:    client.TriggerThresholdOpGreaterThanOrEqual,
				Value: 10000,
			},
			Recipients: []client.NotificationRecipient{
				{
					Type:   client.RecipientTypeMarker,
					Target: "This marker is created by a trigger",
				},
			},
			BaselineDetails: &client.TriggerBaselineDetails{
				Type:          "percentage",
				OffsetMinutes: 1440,
			},
		}
		trigger, err = c.Triggers.Create(ctx, dataset, data)
		require.NoError(t, err)
		assert.NotNil(t, trigger.ID)

		// copy IDs before asserting equality
		data.ID = trigger.ID
		data.QueryID = trigger.QueryID
		data.AlertType = trigger.AlertType
		data.Recipients[0].ID = trigger.Recipients[0].ID
		// set default time range
		data.Query.TimeRange = client.ToPtr(900)

		// set the default alert type
		data.AlertType = client.TriggerAlertTypeOnChange
		// set the default evaluation window type
		data.EvaluationScheduleType = client.TriggerEvaluationScheduleFrequency
		// set the default threshold exceeded limit
		data.Threshold.ExceededLimit = 1
		data.BaselineDetails = &client.TriggerBaselineDetails{
			Type:          "percentage",
			OffsetMinutes: 1440,
		}
		assert.Equal(t, data, trigger)
	})

	t.Run("Get - with baseline details", func(t *testing.T) {
		trigger.BaselineDetails = &client.TriggerBaselineDetails{
			Type:          "percentage",
			OffsetMinutes: 1440,
		}
		getTrigger, err := c.Triggers.Get(ctx, dataset, trigger.ID)

		require.NoError(t, err)
		assert.Equal(t, trigger.BaselineDetails.Type, getTrigger.BaselineDetails.Type)
		assert.Equal(t, trigger.BaselineDetails.OffsetMinutes, getTrigger.BaselineDetails.OffsetMinutes)
	})

	t.Run("Update - with baseline_details", func(t *testing.T) {
		trigger.BaselineDetails = &client.TriggerBaselineDetails{
			Type:          "percentage",
			OffsetMinutes: 1440,
		}
		result, err := c.Triggers.Update(ctx, dataset, trigger)

		require.NoError(t, err)
		assert.Equal(t, trigger.BaselineDetails.Type, result.BaselineDetails.Type)
		assert.Equal(t, trigger.BaselineDetails.OffsetMinutes, result.BaselineDetails.OffsetMinutes)
	})

}
