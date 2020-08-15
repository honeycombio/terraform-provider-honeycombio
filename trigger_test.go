package honeycombio

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
		filterCombinaton := FilterCombinationOr

		data := &Trigger{
			Name:        fmt.Sprintf("Test trigger created at %v", time.Now()),
			Description: "Some description",
			Disabled:    true,
			Query: &QuerySpec{
				Breakdowns: nil,
				Calculations: []CalculationSpec{
					{
						Op:     CalculateOpP99,
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
				FilterCombination: &filterCombinaton,
			},
			Frequency: 300,
			Threshold: &TriggerThreshold{
				Op:    TriggerThresholdOpGreaterThan,
				Value: Float64Ptr(10000),
			},
			Recipients: []TriggerRecipient{
				{
					Type:   TriggerRecipientTypeEmail,
					Target: "hello@example.com",
				},
				{
					Type:   TriggerRecipientTypeMarker,
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
		for i := range trigger.Recipients {
			data.Recipients[i].ID = trigger.Recipients[i].ID
		}

		assert.Equal(t, data, trigger)
	})

	t.Run("List", func(t *testing.T) {
		triggers, err := c.Triggers.List(ctx, dataset)
		if err != nil {
			t.Fatal(err)
		}

		var createdTrigger *Trigger

		for _, tr := range triggers {
			if trigger.ID == tr.ID {
				createdTrigger = &tr
			}
		}
		if createdTrigger == nil {
			t.Fatalf("could not find newly created trigger with ID = %s", trigger.ID)
		}

		assert.Equal(t, *trigger, *createdTrigger)
	})

	t.Run("Update", func(t *testing.T) {
		newTrigger := *trigger
		newTrigger.Disabled = true

		updatedTrigger, err := c.Triggers.Update(ctx, dataset, trigger)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, newTrigger, *updatedTrigger)

		trigger = updatedTrigger
	})

	t.Run("Get", func(t *testing.T) {
		getTrigger, err := c.Triggers.Get(ctx, dataset, trigger.ID)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, *trigger, *getTrigger)
	})

	t.Run("Delete", func(t *testing.T) {
		err = c.Triggers.Delete(ctx, dataset, trigger.ID)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Get_unexistingID", func(t *testing.T) {
		_, err := c.Triggers.Get(ctx, dataset, trigger.ID)
		assert.Equal(t, ErrNotFound, err)
	})

	t.Run("Create_invalid", func(t *testing.T) {
		invalidTrigger := *trigger
		invalidTrigger.ID = ""
		invalidTrigger.Query.Calculations = []CalculationSpec{
			{
				Op: "COUNT",
			},
			{
				Op:     "AVG",
				Column: StringPtr("duration_ms"),
			},
		}

		_, err := c.Triggers.Create(ctx, dataset, &invalidTrigger)
		assert.Equal(t, errors.New("422 Unprocessable Entity: trigger query requires exactly one calculation"), err)
	})
}

func TestMatchesTriggerSubset(t *testing.T) {
	cases := []struct {
		in          QuerySpec
		expectedOk  bool
		expectedErr error
	}{
		{
			in: QuerySpec{
				Calculations: []CalculationSpec{
					{
						Op: CalculateOpCount,
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
						Op: CalculateOpHeatmap,
					},
				},
			},
			expectedErr: errors.New("a trigger query may not contain a HEATMAP calculation"),
		},
		{
			in: QuerySpec{
				Calculations: []CalculationSpec{
					{
						Op: CalculateOpCount,
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
						Op: CalculateOpCount,
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
