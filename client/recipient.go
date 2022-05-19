package client

// BurnAlerts describe all the BurnAlert-related methods that the Honeycomb API supports.
//
// API docs: https://docs.honeycomb.io/api/burn-alerts/
type Recipients interface{}

// recipients implements Recipients.
type recipients struct {
	// client *Client
}

// Compile-time proof of interface implementation by type burnalerts.
var _ Recipients = (*recipients)(nil)

type Recipient struct {
	ID     string `json:"id,omitempty"`
	Type   string `json:"type,omitempty"`
	Target string `json:"target"`
}
