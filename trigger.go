package honeycombio

import (
	"errors"
	"fmt"
)

// Compile-time proof of interface implementation.
var _ Triggers = (*triggers)(nil)

// Triggers describes all the trigger-related methods that Honeycomb supports.
type Triggers interface {
	// List all triggers present in this dataset.
	List(dataset string) ([]Trigger, error)

	// Get a trigger by its ID.
	Get(dataset string, id string) (*Trigger, error)

	// Create a new trigger in this dataset. When creating a new trigger, ID
	// may not be set.
	Create(dataset string, t *Trigger) (*Trigger, error)

	// Update an existing trigger. Missing (optional) fields will set to their
	// respective defaults and not the currently existing values. Except for
	// the disabled flag, which will retain its existing value when omitted.
	Update(dataset string, t *Trigger) (*Trigger, error)

	// Delete a trigger from the dataset.
	Delete(dataset string, id string) error
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

// List of available trigger op types.
const (
	TriggerThresholdOpGreaterThan        TriggerThresholdOp = ">"
	TriggerThresholdOpGreaterThanOrEqual TriggerThresholdOp = ">="
	TriggerThresholdOpLessThan           TriggerThresholdOp = "<"
	TriggerThresholdOpLessThanOrEqual    TriggerThresholdOp = "<="
)

// TriggerRecipient represents a recipient that will receive a notification if
// the trigger fires, as described by https://docs.honeycomb.io/api/triggers/#specifying-recipients
//
// Note: when adding Slack as recipient you have to specify the ID as well.
// This is not supported yet.
type TriggerRecipient struct {
	// Type of the trigger, possible values (not exhaustive) are "email",
	// "slack" and "pagerduty".
	Type string `json:"type"`
	// Target of the trigger, depending on the type this can be an email
	// address or Slack channel.
	Target string `json:"target,omitempty"`

	// TODO add ID
}

func (s *triggers) List(dataset string) ([]Trigger, error) {
	req, err := s.client.newRequest("GET", "/1/triggers/"+urlEncodeDataset(dataset), nil)
	if err != nil {
		return nil, err
	}

	var t []Trigger
	err = s.client.do(req, &t)
	return t, err
}

func (s *triggers) Get(dataset string, id string) (*Trigger, error) {
	req, err := s.client.newRequest("GET", fmt.Sprintf("/1/triggers/%s/%s", urlEncodeDataset(dataset), id), nil)
	if err != nil {
		return nil, err
	}

	var t *Trigger
	err = s.client.do(req, &t)
	return t, err
}

func (s *triggers) Create(dataset string, data *Trigger) (*Trigger, error) {
	if data.ID != "" {
		return nil, errors.New("Trigger.ID must be empty when creating a trigger ")
	}

	req, err := s.client.newRequest("POST", fmt.Sprintf("/1/triggers/%s", urlEncodeDataset(dataset)), data)
	if err != nil {
		return nil, err
	}

	var t *Trigger
	err = s.client.do(req, &t)
	return t, err
}

func (s *triggers) Update(dataset string, data *Trigger) (*Trigger, error) {
	req, err := s.client.newRequest("PUT", fmt.Sprintf("/1/triggers/%s/%s", urlEncodeDataset(dataset), data.ID), data)
	if err != nil {
		return nil, err
	}

	var t *Trigger
	err = s.client.do(req, &t)
	return t, err
}

func (s *triggers) Delete(dataset string, id string) error {
	req, err := s.client.newRequest("DELETE", fmt.Sprintf("/1/triggers/%s/%s", urlEncodeDataset(dataset), id), nil)
	if err != nil {
		return nil
	}

	return s.client.do(req, nil)
}
