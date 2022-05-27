package client

// Recipients creates a Recipient definition for use by BurnAlerts and Triggers in the the Honeycomb API
type Recipient struct {
	ID     string        `json:"id,omitempty"`
	Type   RecipientType `json:"type,omitempty"`
	Target string        `json:"target"`
}

// RecipientType holds all the possible recipient types.
type RecipientType string

// Declaration of recipient types
const (
	RecipientTypeEmail     RecipientType = "email"
	RecipientTypePagerDuty RecipientType = "pagerduty"
	RecipientTypeSlack     RecipientType = "slack"
	RecipientTypeWebhook   RecipientType = "webhook"
	RecipientTypeMarker    RecipientType = "marker"
	RecipientTypeZenoss    RecipientType = "zenoss"
)

// TriggerRecipientTypes returns a list of recipient types compatible with Triggers
func TriggerRecipientTypes() []RecipientType {
	return []RecipientType{
		RecipientTypeEmail,
		RecipientTypePagerDuty,
		RecipientTypeSlack,
		RecipientTypeWebhook,
		RecipientTypeZenoss,
		RecipientTypeMarker,
	}
}

// BurnAlertRecipientTypes returns a list of recipient types compatible with Burn Alerts
func BurnAlertRecipientTypes() []RecipientType {
	return []RecipientType{
		RecipientTypeEmail,
		RecipientTypePagerDuty,
		RecipientTypeSlack,
		RecipientTypeWebhook,
		RecipientTypeZenoss,
	}
}
