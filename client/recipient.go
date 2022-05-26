package client

// Recipients creates a Recipient definition for use by BurnAlerts and Triggers in the the Honeycomb API
type Recipient struct {
	ID     string        `json:"id,omitempty"`
	Type   RecipientType `json:"type,omitempty"`
	Target string        `json:"target"`
}

// RecipientType holds all the possible recipient types.
type RecipientType string

// Declaration of trigger recipient types
const (
	RecipientTypeEmail     RecipientType = "email"
	RecipientTypePagerDuty RecipientType = "pagerduty"
	RecipientTypeSlack     RecipientType = "slack"
	RecipientTypeWebhook   RecipientType = "webhook"
	RecipientTypeZenoss    RecipientType = "zenoss"
)

// RecipientTypes returns an exhaustive list of recipient types.
func RecipientTypes() []RecipientType {
	return []RecipientType{
		RecipientTypeEmail,
		RecipientTypePagerDuty,
		RecipientTypeSlack,
		RecipientTypeWebhook,
		RecipientTypeZenoss,
	}
}
