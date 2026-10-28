package client_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
)

func TestNotificationRecipientDetails_MarshalJSON(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		details  client.NotificationRecipientDetails
		expected string
	}{
		"nil variables are omitted": {
			details:  client.NotificationRecipientDetails{PDSeverity: client.PDSeverityINFO},
			expected: `{"pagerduty_severity":"info"}`,
		},
		"empty variables are sent to clear stored variables": {
			details:  client.NotificationRecipientDetails{Variables: []client.NotificationVariable{}},
			expected: `{"variables":[]}`,
		},
		"variables are sent": {
			details:  client.NotificationRecipientDetails{Variables: []client.NotificationVariable{{Name: "severity", Value: "warning"}}},
			expected: `{"variables":[{"name":"severity","value":"warning"}]}`,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := json.Marshal(client.NotificationRecipient{ID: "abcd1234", Details: &tc.details})
			require.NoError(t, err)
			assert.JSONEq(t, `{"id":"abcd1234","type":"","details":`+tc.expected+`}`, string(got))
		})
	}
}
