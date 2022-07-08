package client

// NotificationRecipient represents a recipient embedded in a Trigger or Burn Alert
type NotificationRecipient struct {
	ID     string        `json:"id"`
	Type   RecipientType `json:"type"`
	Target string        `json:"target,omitempty"`
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
