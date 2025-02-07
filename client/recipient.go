package client

import (
	"context"
	"fmt"
	"time"
)

// Recipients describe all the Recipient-related methods that the Honeycomb API supports.
//
// API docs: https://docs.honeycomb.io/api/recipients/
type Recipients interface {
	// List all Recipients
	List(ctx context.Context) ([]Recipient, error)

	// Get a Recipient by its ID.
	Get(ctx context.Context, id string) (*Recipient, error)

	// Create a new Recipient. When creating a new Recipient ID must not be set.
	Create(ctx context.Context, r *Recipient) (*Recipient, error)

	// Update an existing Recipient.
	Update(ctx context.Context, r *Recipient) (*Recipient, error)

	// Delete a Recipient
	Delete(ctx context.Context, id string) error
}

// recipients implements Recipients
type recipients struct {
	client *Client
}

// Compile-time proof of interface implementation by type recipients.
var _ Recipients = (*recipients)(nil)

// Recipient represents a Honeycomb Recipient
type Recipient struct {
	ID        string           `json:"id,omitempty"`
	Type      RecipientType    `json:"type"`
	Details   RecipientDetails `json:"details"`
	CreatedAt time.Time        `json:"created_at,omitempty"`
	UpdatedAt time.Time        `json:"updated_at,omitempty"`
}

// NotificationRecipient represents a recipient embedded in a Trigger or Burn Alert
type NotificationRecipient struct {
	ID      string                        `json:"id,omitempty"`
	Type    RecipientType                 `json:"type"`
	Details *NotificationRecipientDetails `json:"details,omitempty"`
	Target  string                        `json:"target,omitempty"`
}

type RecipientDetails struct {
	// email
	EmailAddress string `json:"email_address,omitempty"`
	// marker
	MarkerID string `json:"marker_id,omitempty"`
	// pagerduty
	PDIntegrationKey  string `json:"pagerduty_integration_key,omitempty"`
	PDIntegrationName string `json:"pagerduty_integration_name,omitempty"`
	// slack
	SlackChannel string `json:"slack_channel,omitempty"`
	// webhook or msteams
	WebhookName string `json:"webhook_name,omitempty"`
	WebhookURL  string `json:"webhook_url,omitempty"`
	// webhook only
	WebhookSecret  string          `json:"webhook_secret,omitempty"`
	WebhookHeaders []WebhookHeader `json:"webhook_headers"`
	// custom webhook
	WebhookPayloads *WebhookPayloads `json:"webhook_payloads,omitempty"`
}

type NotificationVariable struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type NotificationRecipientDetails struct {
	Variables  []NotificationVariable `json:"variables,omitempty"`
	PDSeverity PagerDutySeverity      `json:"pagerduty_severity,omitempty"`
}

type WebhookPayloads struct {
	PayloadTemplates  PayloadTemplates   `json:"payload_templates"`
	TemplateVariables []TemplateVariable `json:"template_variables"`
}

type PayloadTemplates struct {
	Trigger        *PayloadTemplate `json:"trigger,omitempty"`
	ExhaustionTime *PayloadTemplate `json:"exhaustion_time,omitempty"`
	BudgetRate     *PayloadTemplate `json:"budget_rate,omitempty"`
}

type PayloadTemplate struct {
	Body string `json:"body"`
}

type TemplateVariable struct {
	Name    string `json:"name"`
	Default string `json:"default_value"`
}

type WebhookHeader struct {
	Key   string `json:"header"`
	Value string `json:"value"`
}

// RecipientType holds all the possible recipient types.
type RecipientType string

func (t RecipientType) String() string {
	return string(t)
}

// Declaration of recipient types
const (
	RecipientTypeEmail     RecipientType = "email"
	RecipientTypePagerDuty RecipientType = "pagerduty"
	RecipientTypeSlack     RecipientType = "slack"
	RecipientTypeWebhook   RecipientType = "webhook"
	RecipientTypeMarker    RecipientType = "marker"
	// Deprecated: new recipients must use RecipientTypeMSTeamsWorkflow instead
	RecipientTypeMSTeams         RecipientType = "msteams"
	RecipientTypeMSTeamsWorkflow RecipientType = "msteams_workflow"
)

// PagerDutySeverity holds all the possible PD Severity types
type PagerDutySeverity string

const (
	PDSeverityCRITICAL PagerDutySeverity = "critical"
	PDSeverityERROR    PagerDutySeverity = "error"
	PDSeverityWARNING  PagerDutySeverity = "warning"
	PDSeverityINFO     PagerDutySeverity = "info"
	PDDefaultSeverity                    = PDSeverityCRITICAL
)

// TriggerRecipientTypes returns a list of recipient types compatible with Triggers.
// Triggers are a special case as 'Marker' recipients are supported in addition to
// usual types.
func TriggerRecipientTypes() []RecipientType {
	return append(RecipientTypes(), RecipientTypeMarker)
}

// RecipientTypes returns all supported Recipient types
func RecipientTypes() []RecipientType {
	return []RecipientType{
		RecipientTypeEmail,
		RecipientTypePagerDuty,
		RecipientTypeSlack,
		RecipientTypeWebhook,
		RecipientTypeMSTeams,
		RecipientTypeMSTeamsWorkflow,
	}
}

func (s *recipients) List(ctx context.Context) ([]Recipient, error) {
	var r []Recipient
	err := s.client.Do(ctx, "GET", "/1/recipients", nil, &r)
	return r, err
}

func (s *recipients) Get(ctx context.Context, ID string) (*Recipient, error) {
	var r Recipient
	err := s.client.Do(ctx, "GET", fmt.Sprintf("/1/recipients/%s", ID), nil, &r)
	return &r, err
}

func (s *recipients) Create(ctx context.Context, data *Recipient) (*Recipient, error) {
	var r Recipient
	err := s.client.Do(ctx, "POST", "/1/recipients", data, &r)
	return &r, err
}

func (s *recipients) Update(ctx context.Context, data *Recipient) (*Recipient, error) {
	var r Recipient
	err := s.client.Do(ctx, "PUT", fmt.Sprintf("/1/recipients/%s", data.ID), data, &r)
	return &r, err
}

func (s *recipients) Delete(ctx context.Context, id string) error {
	return s.client.Do(ctx, "DELETE", fmt.Sprintf("/1/recipients/%s", id), nil, nil)
}
