package client

// Recipients creates a Recipient definition for use by BurnAlerts and Triggers in the the Honeycomb API
//
// API docs: https://docs.honeycomb.io/api/burn-alerts/
type Recipients interface{}
type Recipient struct {
	ID     string `json:"id,omitempty"`
	Type   string `json:"type,omitempty"`
	Target string `json:"target"`
}
