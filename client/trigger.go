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

	// Get a trigger by its ID. Returns ErrNotFound if there is no trigger with
	// the given ID in this dataset.
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
	Query   *QuerySpec `json:"query,omitempty"`
	QueryID string     `json:"query_id,omitempty"`
	// Alert Frequency. Describes scheduling behavior for triggers. By default alert type per change. This field is required
	AlertType string `json:"alert_type,omitempty"`
	// Threshold. This fild is required.
	Threshold *TriggerThreshold `json:"threshold"`
	// Frequency describes how often the trigger should run. Frequency is an
	// interval in seconds, defaulting to 900 (15 minutes). Its value must be
	// divisible by 60 and between 60 and 86400 (between 1 minute and 1 day).
	Frequency int `json:"frequency,omitempty"`
	// Recipients are notified when the trigger fires.
	Recipients []TriggerRecipient `json:"recipients,omitempty"`
}

// TriggerThreshold represents the threshold of a trigger.
type TriggerThreshold struct {
	Op    TriggerThresholdOp `json:"op"`
	Value float64            `json:"value"`
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

// Allowed values for alert_type. | on_change is default
const (
	TriggerAlertTypeValueOnChange string = "on_change"
	TriggerAlertTypeValueOnTrue   string = "on_true"
)

// TriggerRecipient represents a recipient that will receive a notification
// when the trigger fires.
//
// API docs: https://docs.honeycomb.io/api/triggers/#specifying-recipients
//
// Notes
//
// Recipients of type Slack should be specified by their ID. It is not possible
// to create a new recipient of type Slack using the API. Instead use the ID of
// a recipient of type Slack that was manually added to another trigger.
//
// Recipients of type webhook can be added by their name. If a webhook with
// this name does not exist yet (or if the name contains a typo), the Honeycomb
// API will not complain about this but the webhook will not be valid.
type TriggerRecipient struct {
	// ID of the recipient.
	ID string `json:"id,omitempty"`
	// Type of the recipient.
	Type TriggerRecipientType `json:"type"`
	// Target of the trigger, this has another meaning depending on type:
	// - email: an email address
	// - marker: name of the marker
	// - PagerDuty: N/A
	// - Slack: name of a channel
	// - Webhook: name of the webhook
	Target string `json:"target,omitempty"`
}

// Custom JSON marhsal to work around https://github.com/honeycombio/terraform-provider-honeycombio/issues/123
//
// It seems the Plugin SDK may reuse the value of 'target' for another
// member of the list if target is unset (null or empty string)
func (tr *TriggerRecipient) MarshalJSON() ([]byte, error) {
	// aliased type to avoid stack overflows due to recursion
	type ARecipient TriggerRecipient

	if tr.Type == TriggerRecipientTypePagerDuty {
		r := &ARecipient{
			ID:   tr.ID,
			Type: tr.Type,
		}
		return json.Marshal(&struct{ *ARecipient }{ARecipient: (*ARecipient)(r)})
	}

	return json.Marshal(&struct{ *ARecipient }{ARecipient: (*ARecipient)(tr)})
}

// TriggerRecipientType holds all the possible trigger recipient types.
type TriggerRecipientType string

// Declaration of trigger recipient types
const (
	TriggerRecipientTypeEmail     TriggerRecipientType = "email"
	TriggerRecipientTypeMarker    TriggerRecipientType = "marker"
	TriggerRecipientTypePagerDuty TriggerRecipientType = "pagerduty"
	TriggerRecipientTypeSlack     TriggerRecipientType = "slack"
	TriggerRecipientTypeWebhook   TriggerRecipientType = "webhook"
)

// TriggerRecipientTypes returns an exhaustive list of trigger recipient types.
func TriggerRecipientTypes() []TriggerRecipientType {
	return []TriggerRecipientType{
		TriggerRecipientTypeEmail,
		TriggerRecipientTypeMarker,
		TriggerRecipientTypePagerDuty,
		TriggerRecipientTypeSlack,
		TriggerRecipientTypeWebhook,
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
			ID:          t.ID,
			Name:        t.Name,
			Description: t.Description,
			Disabled:    t.Disabled,
			QueryID:     t.QueryID,
			AlertType:   t.AlertType,
			Threshold:   t.Threshold,
			Frequency:   t.Frequency,
			Recipients:  t.Recipients,
		}
		return json.Marshal(&struct{ *ATrigger }{ATrigger: (*ATrigger)(a)})
	}

	return json.Marshal(&struct{ *ATrigger }{ATrigger: (*ATrigger)(t)})
}

func (s *triggers) List(ctx context.Context, dataset string) ([]Trigger, error) {
	var t []Trigger
	err := s.client.performRequest(ctx, "GET", "/1/triggers/"+urlEncodeDataset(dataset), nil, &t)
	return t, err
}

func (s *triggers) Get(ctx context.Context, dataset string, id string) (*Trigger, error) {
	var t Trigger
	err := s.client.performRequest(ctx, "GET", fmt.Sprintf("/1/triggers/%s/%s", urlEncodeDataset(dataset), id), nil, &t)
	return &t, err
}

func (s *triggers) Create(ctx context.Context, dataset string, data *Trigger) (*Trigger, error) {
	var t Trigger
	err := s.client.performRequest(ctx, "POST", fmt.Sprintf("/1/triggers/%s", urlEncodeDataset(dataset)), data, &t)
	return &t, err
}

func (s *triggers) Update(ctx context.Context, dataset string, data *Trigger) (*Trigger, error) {
	var t Trigger
	err := s.client.performRequest(ctx, "PUT", fmt.Sprintf("/1/triggers/%s/%s", urlEncodeDataset(dataset), data.ID), data, &t)
	return &t, err
}

func (s *triggers) Delete(ctx context.Context, dataset string, id string) error {
	return s.client.performRequest(ctx, "DELETE", fmt.Sprintf("/1/triggers/%s/%s", urlEncodeDataset(dataset), id), nil, nil)
}

// MatchesTriggerSubset checks that the given QuerySpec matches the strict
// subset required to be used in a trigger.
//
// The following properties must be valid:
//
//  - the query must contain exactly one calculation
//  - the HEATMAP calculation may not be used
//  - only the following fields may be set: calculations, breakdown, filters, filter_combination and time_range
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
