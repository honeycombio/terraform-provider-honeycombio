package honeycombio

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTriggers(t *testing.T) {
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
						Column: &[]string{"duration_ms"}[0],
					},
				},
				Filters: []FilterSpec{
					{
						Column: "trace.parent_id",
						Op:     FilterOpExists,
					},
					{
						Column: "github.ref",
						Op:     FilterOpContains,
						Value:  "main",
					},
				},
				FilterCombination: &filterCombinaton,
			},
			Frequency: 300,
			Threshold: &TriggerThreshold{
				Op:    TriggerThresholdOpGreaterThan,
				Value: &[]float64{10000}[0],
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
		trigger, err = c.Triggers.Create(dataset, data)
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
		triggers, err := c.Triggers.List(dataset)
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

		updatedTrigger, err := c.Triggers.Update(dataset, trigger)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, newTrigger, *updatedTrigger)

		trigger = updatedTrigger
	})

	t.Run("Get", func(t *testing.T) {
		getTrigger, err := c.Triggers.Get(dataset, trigger.ID)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, *trigger, *getTrigger)
	})

	t.Run("Delete", func(t *testing.T) {
		err = c.Triggers.Delete(dataset, trigger.ID)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Get_unexistingID", func(t *testing.T) {
		_, err := c.Triggers.Get(dataset, trigger.ID)
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
				Column: &[]string{"duration_ms"}[0],
			},
		}

		_, err := c.Triggers.Create(dataset, &invalidTrigger)
		assert.Equal(t, errors.New("request failed (422): trigger query requires exactly one calculation"), err)
	})
}
