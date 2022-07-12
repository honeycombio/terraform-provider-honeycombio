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
						Column: StringPtr("duration_ms"),
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
					Type:   RecipientTypeEmail,
					Target: "hello@example.com",
				},
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
		for i := range trigger.Recipients {
			data.Recipients[i].ID = trigger.Recipients[i].ID
		}
		// set default time range
		data.Query.TimeRange = IntPtr(300)

		// set the default alert type
		data.AlertType = TriggerAlertTypeValueOnChange

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

		// update the alert_type to on_true / this is a simple validation that this field works
		trigger.AlertType = TriggerAlertTypeValueOnTrue

		result, err := c.Triggers.Update(ctx, dataset, trigger)

		// copy IDs before asserting equality
		trigger.QueryID = result.QueryID
		trigger.AlertType = result.AlertType
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
				Limit: IntPtr(100),
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
						Column: StringPtr("duration_ms"),
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
