package client

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTriggers(t *testing.T) {
	ctx := context.Background()

	var trigger *Trigger
	var err error

	c := newTestClient(t)
	dataset := testDataset(t)

	t.Run("Create", func(t *testing.T) {
		data := &Trigger{
			Name:        fmt.Sprintf("Test trigger created at %v", time.Now()),
			Description: "Some description",
			Disabled:    true,
			Query: &QuerySpec{
				Breakdowns: nil,
				Calculations: []CalculationSpec{
					{
						Op:     CalculationOpP99,
						Column: ToPtr("duration_ms"),
					},
				},
				Filters: []FilterSpec{
					{
						Column: "column_1",
						Op:     FilterOpExists,
					},
					{
						Column: "column_2",
						Op:     FilterOpContains,
						Value:  "foobar",
					},
				},
				FilterCombination: FilterCombinationOr,
			},
			Frequency: 300,
			Threshold: &TriggerThreshold{
				Op:    TriggerThresholdOpGreaterThan,
				Value: 10000,
			},
			Recipients: []NotificationRecipient{
				{
					Type:   RecipientTypeMarker,
					Target: "This marker is created by a trigger",
				},
			},
		}
		trigger, err = c.Triggers.Create(ctx, dataset, data)

		if err != nil {
			t.Fatal(err)
		}
		assert.NotNil(t, trigger.ID)

		// copy IDs before asserting equality
		data.ID = trigger.ID
		data.QueryID = trigger.QueryID
		data.AlertType = trigger.AlertType
		data.Recipients[0].ID = trigger.Recipients[0].ID
		// set default time range
		data.Query.TimeRange = ToPtr(300)

		// set the default alert type
		data.AlertType = TriggerAlertTypeOnChange
		// set the default evaluation window type
		data.EvaluationScheduleType = TriggerEvaluationScheduleFrequency
		// set the default threshold exceeded limit
		data.Threshold.ExceededLimit = 1

		assert.Equal(t, data, trigger)
	})

	t.Run("List", func(t *testing.T) {
		result, err := c.Triggers.List(ctx, dataset)

		assert.NoError(t, err)
		assert.Contains(t, result, *trigger, "could not find newly created trigger with List")
	})

	t.Run("Get", func(t *testing.T) {
		getTrigger, err := c.Triggers.Get(ctx, dataset, trigger.ID)

		assert.NoError(t, err)
		assert.Equal(t, *trigger, *getTrigger)
	})

	t.Run("Update", func(t *testing.T) {
		trigger.Description = "A new description"

		// update the alert type to on true
		trigger.AlertType = TriggerAlertTypeOnTrue
		// update the evaluation schedule to a window type
		trigger.EvaluationScheduleType = TriggerEvaluationScheduleWindow
		trigger.EvaluationSchedule = &TriggerEvaluationSchedule{
			Window: TriggerEvaluationWindow{
				DaysOfWeek: []string{"monday", "wednesday", "friday"},
				StartTime:  "13:00",
				EndTime:    "21:00",
			},
		}
		// update the threshold exceeded limit to 3
		trigger.Threshold.ExceededLimit = 3

		result, err := c.Triggers.Update(ctx, dataset, trigger)

		// copy IDs before asserting equality
		trigger.QueryID = result.QueryID
		assert.NoError(t, err)
		assert.Equal(t, trigger, result)
	})

	t.Run("Delete", func(t *testing.T) {
		err = c.Triggers.Delete(ctx, dataset, trigger.ID)

		assert.NoError(t, err)
	})

	t.Run("Get_deletedTrigger", func(t *testing.T) {
		_, err := c.Triggers.Get(ctx, dataset, trigger.ID)

		assert.Equal(t, ErrNotFound, err)
	})
}

func TestMatchesTriggerSubset(t *testing.T) {
	cases := []struct {
		in          QuerySpec
		expectedErr error
	}{
		{
			in: QuerySpec{
				Calculations: []CalculationSpec{
					{
						Op: CalculationOpCount,
					},
				},
			},
			expectedErr: nil,
		},
		{
			in: QuerySpec{
				Calculations: nil,
			},
			expectedErr: errors.New("a trigger query should contain exactly one calculation"),
		},
		{
			in: QuerySpec{
				Calculations: []CalculationSpec{
					{
						Op: CalculationOpHeatmap,
					},
				},
			},
			expectedErr: errors.New("a trigger query may not contain a HEATMAP calculation"),
		},
		{
			in: QuerySpec{
				Calculations: []CalculationSpec{
					{
						Op: CalculationOpCount,
					},
				},
				Limit: ToPtr(100),
			},
			expectedErr: errors.New("limit is not allowed in a trigger query"),
		},
		{
			in: QuerySpec{
				Calculations: []CalculationSpec{
					{
						Op: CalculationOpCount,
					},
				},
				Orders: []OrderSpec{
					{
						Column: ToPtr("duration_ms"),
					},
				},
			},
			expectedErr: errors.New("orders is not allowed in a trigger query"),
		},
	}

	for i, c := range cases {
		err := MatchesTriggerSubset(&c.in)

		assert.Equal(t, c.expectedErr, err, "Test case %d, QuerySpec: %v", i, c.in)
	}
}
