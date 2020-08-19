package honeycombio

import (
	"context"
	"errors"
	"fmt"
)

// Compile-time proof of interface implementation.
var _ Triggers = (*triggers)(nil)

// Triggers describes all the trigger-related methods that Honeycomb supports.
type Triggers interface {
	// List all triggers present in this dataset.
	List(ctx context.Context, dataset string) ([]Trigger, error)

	// Get a trigger by its ID.
	Get(ctx context.Context, dataset string, id string) (*Trigger, error)

	// Create a new trigger in this dataset. When creating a new trigger, ID
	// may not be set.
	Create(ctx context.Context, dataset string, t *Trigger) (*Trigger, error)

	// Update an existing trigger. Missing (optional) fields will set to their
	// respective defaults and not the currently existing values. Except for
	// the disabled flag, which will retain its existing value when omitted.
	Update(ctx context.Context, dataset string, t *Trigger) (*Trigger, error)

	// Delete a trigger from the dataset.
	Delete(ctx context.Context, dataset string, id string) error
}

// trigger implements Triggers.
type triggers struct {
	client *Client
}

// Trigger represents a Honeycomb trigger, as described by https://docs.honeycomb.io/api/triggers/#fields-on-a-trigger
type Trigger struct {
	ID string `json:"id,omitempty"`
	// Name of the trigger, required when creating a new trigger.
	Name        string `json:"name,omitempty"`
	Disabled    bool   `json:"disabled,omitempty"`
	Description string `json:"description,omitempty"`
	// Query of the trigger, required when creating a new trigger. The query of
	// a trigger must contain exactly one item in `calculations`. The HEATMAP
	// calculation may not be used.
	Query *QuerySpec `json:"query,omitempty"`
	// Frequency as an interval in seconds, defaults to 900 (15 minutes). Value
	// must be divisible by 60 and between 60 and 86400 (between 1 minute and 1
	// day).
	Frequency int `json:"frequency,omitempty"`
	// Threshold, required when creating a new trigger.
	Threshold  *TriggerThreshold  `json:"threshold,omitempty"`
	Recipients []TriggerRecipient `json:"recipients,omitempty"`
}

// TriggerThreshold represents the threshold of a trigger.
type TriggerThreshold struct {
	Op    TriggerThresholdOp `json:"op"`
	Value *float64           `json:"value"`
}

// TriggerThresholdOp the operator of the trigger threshold.
type TriggerThresholdOp string

// Declaration of trigger threshold ops.
const (
	TriggerThresholdOpGreaterThan        TriggerThresholdOp = ">"
	TriggerThresholdOpGreaterThanOrEqual TriggerThresholdOp = ">="
	TriggerThresholdOpLessThan           TriggerThresholdOp = "<"
	TriggerThresholdOpLessThanOrEqual    TriggerThresholdOp = "<="
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

// TriggerRecipient represents a recipient that will receive a notification if
// the trigger fires, as described by https://docs.honeycomb.io/api/triggers/#specifying-recipients
type TriggerRecipient struct {
	// ID of the recipient, this is required when type is Slack.
	ID string `json:"id,omitempty"`
	// Type of the recipient.
	Type TriggerRecipientType `json:"type"`
	// Target of the trigger, this has another meaning depending on type:
	// - email: an email address
	// - marker: name of the marker
	// - PagerDuty: N/A
	// - Slack: name of a channel
	Target string `json:"target,omitempty"`
}

// TriggerRecipientType holds all the possible trigger recipient types.
type TriggerRecipientType string

// Declaration of trigger recipient types
const (
	TriggerRecipientTypeEmail     TriggerRecipientType = "email"
	TriggerRecipientTypeMarker    TriggerRecipientType = "marker"
	TriggerRecipientTypePagerDuty TriggerRecipientType = "pagerduty"
	TriggerRecipientTypeSlack     TriggerRecipientType = "slack"
)

// TriggerRecipientTypes returns an exhaustive list of trigger recipient types.
func TriggerRecipientTypes() []TriggerRecipientType {
	return []TriggerRecipientType{
		TriggerRecipientTypeEmail,
		TriggerRecipientTypeMarker,
		TriggerRecipientTypePagerDuty,
		TriggerRecipientTypeSlack,
	}
}

func (s *triggers) List(ctx context.Context, dataset string) ([]Trigger, error) {
	req, err := s.client.newRequest(ctx, "GET", "/1/triggers/"+urlEncodeDataset(dataset), nil)
	if err != nil {
		return nil, err
	}

	var t []Trigger
	err = s.client.do(req, &t)
	return t, err
}

func (s *triggers) Get(ctx context.Context, dataset string, id string) (*Trigger, error) {
	req, err := s.client.newRequest(ctx, "GET", fmt.Sprintf("/1/triggers/%s/%s", urlEncodeDataset(dataset), id), nil)
	if err != nil {
		return nil, err
	}

	var t *Trigger
	err = s.client.do(req, &t)
	return t, err
}

func (s *triggers) Create(ctx context.Context, dataset string, data *Trigger) (*Trigger, error) {
	if data.ID != "" {
		return nil, errors.New("Trigger.ID must be empty when creating a trigger ")
	}

	req, err := s.client.newRequest(ctx, "POST", fmt.Sprintf("/1/triggers/%s", urlEncodeDataset(dataset)), data)
	if err != nil {
		return nil, err
	}

	var t *Trigger
	err = s.client.do(req, &t)
	return t, err
}

func (s *triggers) Update(ctx context.Context, dataset string, data *Trigger) (*Trigger, error) {
	req, err := s.client.newRequest(ctx, "PUT", fmt.Sprintf("/1/triggers/%s/%s", urlEncodeDataset(dataset), data.ID), data)
	if err != nil {
		return nil, err
	}

	var t *Trigger
	err = s.client.do(req, &t)
	return t, err
}

func (s *triggers) Delete(ctx context.Context, dataset string, id string) error {
	req, err := s.client.newRequest(ctx, "DELETE", fmt.Sprintf("/1/triggers/%s/%s", urlEncodeDataset(dataset), id), nil)
	if err != nil {
		return nil
	}

	return s.client.do(req, nil)
}

// MatchesTriggerSubset checks that the given QuerySpec matches the strict
// subset required to be used in a trigger.
//
// The following properties must be valid:
//
// - the query must contain exactly one calculation
// - the HEATMAP calculation may not be used
// - only the following fields may be set: calculations, breakdown, filters, filter_combination and time_range
//
// For more information, refer to https://docs.honeycomb.io/api/triggers/#fields-on-a-trigger
func MatchesTriggerSubset(query *QuerySpec) error {
	if len(query.Calculations) != 1 {
		return errors.New("a trigger query should contain exactly one calculation")
	}

	if query.Calculations[0].Op == CalculateOpHeatmap {
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
