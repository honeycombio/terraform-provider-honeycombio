package helper

import "github.com/honeycombio/terraform-provider-honeycombio/client"

func TriggerThresholdOpStrings() []string {
	in := client.TriggerThresholdOps()
	out := make([]string, len(in))

	for i := range in {
		out[i] = string(in[i])
	}

	return out
}

func RecipientTypeStrings(recipientTypes []client.RecipientType) []string {
	in := recipientTypes
	out := make([]string, len(in))

	for i := range in {
		out[i] = string(in[i])
	}

	return out
}
