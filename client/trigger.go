package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

// Triggers describes all the trigger-related methods that the Honeycomb API
// supports.
//
// API docs: https://docs.honeycomb.io/api/triggers/
type Triggers interface {
	// List all triggers present in this dataset.
	List(ctx context.Context, dataset string) ([]Trigger, error)

	// Get a trigger by its ID.
	Get(ctx context.Context, dataset string, id string) (*Trigger, error)

	// Create a new trigger in this dataset. When creating a new trigger ID
	// may not be set.
	Create(ctx context.Context, dataset string, t *Trigger) (*Trigger, error)

	// Update an existing trigger. Missing (optional) fields will set to their
	// respective defaults and not the currently existing values. Except for
	// the disabled flag, which will retain its existing value when omitted.
	Update(ctx context.Context, dataset string, t *Trigger) (*Trigger, error)

	// Delete a trigger from the dataset.
	Delete(ctx context.Context, dataset string, id string) error
}

// triggers implements Triggers.
type triggers struct {
	client *Client
}

// Compile-time proof of interface implementation by type triggers.
var _ Triggers = (*triggers)(nil)

// Trigger represents a Honeycomb trigger.
//
// API docs: https://docs.honeycomb.io/api/triggers/#fields-on-a-trigger
type Trigger struct {
	ID string `json:"id,omitempty"`

	// Name of the trigger. This field is required.
	Name string `json:"name"`
	// Description is displayed on the triggers page.
	Description string `json:"description,omitempty"`
	// State of the trigger, if disabled is true the trigger will not run.
	Disabled bool `json:"disabled"`
	// Query of the trigger. This field is required. The query must respect the
	// properties described with and validated by MatchesTriggerSubset.
	// Additionally, time_range of the query can be at most 1 day and may not
	// be greater than 4 times the frequency.
	Query *QuerySpec `json:"query,omitempty"`
	// The ID of the Query of the Trigger. Conflicts with Query
	QueryID string `json:"query_id,omitempty"`
	// Alert Type. Describes scheduling behavior for triggers.
	// Defaults to "on_change"
	AlertType TriggerAlertType `json:"alert_type,omitempty"`
	// Threshold. This fild is required.
	Threshold *TriggerThreshold `json:"threshold"`
	// Evaluation Schedule used by the trigger.
	// Defaults to "frequency", where the trigger runs at the specified frequency.
	// The "window" type means that the trigger will run at the specified frequency,
	// but only in the time window specified by the evaluation schedule.
	EvaluationScheduleType TriggerEvaluationScheduleType `json:"evaluation_schedule_type,omitempty"`
	EvaluationSchedule     *TriggerEvaluationSchedule    `json:"evaluation_schedule,omitempty"`
	// Frequency describes how often the trigger should run. Frequency is an
	// interval in seconds, defaulting to 900 (15 minutes). Its value must be
	// divisible by 60 and between 60 and 86400 (between 1 minute and 1 day).
	Frequency int `json:"frequency,omitempty"`
	// Recipients are notified when the trigger fires.
	Recipients      []NotificationRecipient `json:"recipients,omitempty"`
	BaselineDetails *TriggerBaselineDetails `json:"baseline_details,omitempty"`
	// Tags are used to categorize triggers. They can be used to filtering triggers
	// and are useful for grouping triggers together.
	Tags []Tag `json:"tags"`
}

type TriggerBaselineDetails struct {
	Type          string `json:"type"`
	OffsetMinutes int    `json:"offset_minutes"`
}

// TriggerThreshold represents the threshold of a trigger.
type TriggerThreshold struct {
	Op            TriggerThresholdOp `json:"op"`
	Value         float64            `json:"value"`
	ExceededLimit int                `json:"exceeded_limit,omitempty"`
}

// TriggerThresholdOp the operator of the trigger threshold.
type TriggerThresholdOp string

// TriggerAlertType determines the alert type of a trigger. Valid values are 'on_change' or 'on_true'
type TriggerAlertType string

// TriggerEvaluationScheduleType determines the evaluation schedule type of a trigger. Valid values are 'frequency' or 'window'
type TriggerEvaluationScheduleType string

type TriggerEvaluationSchedule struct {
	Window TriggerEvaluationWindow `json:"window"`
}

type TriggerEvaluationWindow struct {
	// A slice of the days of the week to evaluate the trigger on. (e.g. ["monday", "tuesday", "wednesday"])
	DaysOfWeek []string `json:"days_of_week"`
	// UTC time in HH:mm format (e.g. 13:00)
	StartTime string `json:"start_time"`
	// UTC time in HH:mm format (e.g. 13:00)
	EndTime string `json:"end_time"`
}

const (
	// Trigger threshold ops
	TriggerThresholdOpGreaterThan        TriggerThresholdOp = ">"
	TriggerThresholdOpGreaterThanOrEqual TriggerThresholdOp = ">="
	TriggerThresholdOpLessThan           TriggerThresholdOp = "<"
	TriggerThresholdOpLessThanOrEqual    TriggerThresholdOp = "<="
	// Trigger alert types
	TriggerAlertTypeOnChange TriggerAlertType = "on_change"
	TriggerAlertTypeOnTrue   TriggerAlertType = "on_true"
	// Trigger evaluation schedule types
	TriggerEvaluationScheduleFrequency TriggerEvaluationScheduleType = "frequency"
	TriggerEvaluationScheduleWindow    TriggerEvaluationScheduleType = "window"
)

// TriggerThresholdOps returns an exhaustive list of trigger threshold ops.
func TriggerThresholdOps() []TriggerThresholdOp {
	return []TriggerThresholdOp{
		TriggerThresholdOpGreaterThan,
		TriggerThresholdOpGreaterThanOrEqual,
		TriggerThresholdOpLessThan,
		TriggerThresholdOpLessThanOrEqual,
	}
}

func (t *Trigger) MarshalJSON() ([]byte, error) {
	// aliased type to avoid stack overflows due to recursion
	type ATrigger Trigger

	if t.QueryID != "" && t.Query != nil {
		// we can't sent both to the API, so favour QueryID
		// this doesn't work in the general case, but this
		// client is now purpose-built for the Terraform provider
		a := &ATrigger{
			ID:                     t.ID,
			Name:                   t.Name,
			Description:            t.Description,
			Disabled:               t.Disabled,
			QueryID:                t.QueryID,
			AlertType:              t.AlertType,
			Threshold:              t.Threshold,
			Frequency:              t.Frequency,
			Recipients:             t.Recipients,
			EvaluationScheduleType: t.EvaluationScheduleType,
			EvaluationSchedule:     t.EvaluationSchedule,
		}
		return json.Marshal(&struct{ *ATrigger }{ATrigger: a})
	}

	return json.Marshal(&struct{ *ATrigger }{ATrigger: (*ATrigger)(t)})
}

func (s *triggers) List(ctx context.Context, dataset string) ([]Trigger, error) {
	var t []Trigger
	err := s.client.Do(ctx, "GET", "/1/triggers/"+urlEncodeDataset(dataset), nil, &t)
	return t, err
}

func (s *triggers) Get(ctx context.Context, dataset string, id string) (*Trigger, error) {
	var t Trigger
	err := s.client.Do(ctx, "GET", fmt.Sprintf("/1/triggers/%s/%s", urlEncodeDataset(dataset), id), nil, &t)
	return &t, err
}

func (s *triggers) Create(ctx context.Context, dataset string, data *Trigger) (*Trigger, error) {
	var t Trigger
	err := s.client.Do(ctx, "POST", fmt.Sprintf("/1/triggers/%s", urlEncodeDataset(dataset)), data, &t)
	return &t, err
}

func (s *triggers) Update(ctx context.Context, dataset string, data *Trigger) (*Trigger, error) {
	var t Trigger
	err := s.client.Do(ctx, "PUT", fmt.Sprintf("/1/triggers/%s/%s", urlEncodeDataset(dataset), data.ID), data, &t)
	return &t, err
}

func (s *triggers) Delete(ctx context.Context, dataset string, id string) error {
	return s.client.Do(ctx, "DELETE", fmt.Sprintf("/1/triggers/%s/%s", urlEncodeDataset(dataset), id), nil, nil)
}

// MatchesTriggerSubset checks that the given QuerySpec matches the strict
// subset required to be used in a trigger.
//
// The following properties must be valid:
//
//   - the query must contain exactly one calculation
//   - the HEATMAP calculation may not be used
//   - only the following fields may be set: calculations, breakdown, filters, filter_combination and time_range
//
// For more information, refer to https://docs.honeycomb.io/api/triggers/#fields-on-a-trigger
func MatchesTriggerSubset(query *QuerySpec) error {
	if len(query.Calculations) != 1 {
		return errors.New("a trigger query should contain exactly one calculation")
	}

	if query.Calculations[0].Op == CalculationOpHeatmap {
		return errors.New("a trigger query may not contain a HEATMAP calculation")
	}

	if query.Orders != nil {
		return errors.New("orders is not allowed in a trigger query")
	}

	if query.Limit != nil {
		return errors.New("limit is not allowed in a trigger query")
	}

	return nil
}
