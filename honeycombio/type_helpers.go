package honeycombio

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
)

func boardStyleStrings() []string {
	in := honeycombio.BoardStyles()
	out := make([]string, len(in))

	for i := range in {
		out[i] = string(in[i])
	}

	return out
}

func boardQueryStyleStrings() []string {
	in := honeycombio.BoardQueryStyles()
	out := make([]string, len(in))

	for i := range in {
		out[i] = string(in[i])
	}

	return out
}

func calculationOpStrings() []string {
	in := honeycombio.CalculationOps()
	out := make([]string, len(in))

	for i := range in {
		out[i] = string(in[i])
	}

	return out
}

func columnTypeStrings() []string {
	in := honeycombio.ColumnTypes()
	out := make([]string, len(in))

	for i := range in {
		out[i] = string(in[i])
	}

	return out
}

func filterOpStrings() []string {
	in := honeycombio.FilterOps()
	out := make([]string, len(in))

	for i := range in {
		out[i] = string(in[i])
	}

	return out
}

func havingOpStrings() []string {
	in := honeycombio.HavingOps()
	out := make([]string, len(in))

	for i := range in {
		out[i] = string(in[i])
	}

	return out
}

func havingCalculateOpStrings() []string {
	in := honeycombio.CalculationOps()
	out := make([]string, len(in))

	for i := range in {
		// havings cannot use HEATMAP
		if in[i] != honeycombio.CalculationOpHeatmap {
			out[i] = string(in[i])
		}
	}

	return out
}

func sortOrderStrings() []string {
	in := honeycombio.SortOrders()
	out := make([]string, len(in))

	for i := range in {
		out[i] = string(in[i])
	}

	return out
}

func recipientTypeStrings(recipientTypes []honeycombio.RecipientType) []string {
	in := recipientTypes
	out := make([]string, len(in))

	for i := range in {
		out[i] = string(in[i])
	}

	return out
}

func triggerThresholdOpStrings() []string {
	in := honeycombio.TriggerThresholdOps()
	out := make([]string, len(in))

	for i := range in {
		out[i] = string(in[i])
	}

	return out
}

func coerceValueToType(i string) interface{} {
	// HCL really has three base types: bool, string, and number
	// The Plugin SDK allows typing a schema field to Int or Float

	// Plugin SDK assumes 64bit so we'll do the same
	if v, err := strconv.ParseInt(i, 10, 64); err == nil {
		return int(v)
	} else if v, err := strconv.ParseFloat(i, 64); err == nil {
		return v
	} else if v, err := strconv.ParseBool(i); err == nil {
		return v
	}
	// fallthrough to string
	return i
}

// The SLO API uses 'Target Per Million' to avoid the problems with floats.
// In the name of offering a nicer UX with percentages, we handle the conversion
// back and fourth to allow things like 99.98 to be provided in the HCL and
// handle the conversion to and from 999800

// converts a floating point percentage to a 'Target Per Million' SLO value
func floatToTPM(f float64) int {
	return int(f * 10000)
}

// converts a SLO 'Target Per Million' value to a floating point percentage
func tpmToFloat(t int) float64 {
	return float64(t) / 10000
}

func flattenNotificationRecipients(rs []honeycombio.NotificationRecipient) []map[string]interface{} {
	result := make([]map[string]interface{}, len(rs))

	for i, r := range rs {
		result[i] = map[string]interface{}{
			"id":     r.ID,
			"type":   string(r.Type),
			"target": r.Target,
		}
	}

	return result
}

func expandNotificationRecipients(s []interface{}) []honeycombio.NotificationRecipient {
	recipients := make([]honeycombio.NotificationRecipient, len(s))

	for i, r := range s {
		rMap := r.(map[string]interface{})

		recipients[i] = honeycombio.NotificationRecipient{
			ID:     rMap["id"].(string),
			Type:   honeycombio.RecipientType(rMap["type"].(string)),
			Target: rMap["target"].(string),
		}
	}

	return recipients
}

// Matches read recipients against those declared in HCL and returns
// the Trigger recipients in a stable order grouped by recipient type.
//
// This cannot currently be handled efficiently by a DiffSuppressFunc.
// See: https://github.com/hashicorp/terraform-plugin-sdk/issues/477
func matchNotificationRecipientsWithSchema(readRecipients []honeycombio.NotificationRecipient, declaredRecipients []interface{}) []honeycombio.NotificationRecipient {
	result := make([]honeycombio.NotificationRecipient, len(declaredRecipients))

	rMap := make(map[string]honeycombio.NotificationRecipient, len(readRecipients))
	for _, recipient := range readRecipients {
		rMap[recipient.ID] = recipient
	}

	// Build up result, with each readRecipient in the same position as it
	// appears in declaredRecipients, by looking at each declaredRecipient and
	// finding its matching readRecipient (via rMap).
	//
	// If the declaredRecipient has an ID, this is easy: just look it up and
	// put it in it's place. Otherwise, try to match it to a readRecipient with
	// the same type and target. If we can't find it at all, it must be new, so
	// put it at the end.
	for i, declaredRcpt := range declaredRecipients {
		declaredRcpt := declaredRcpt.(map[string]interface{})

		if declaredRcpt["id"] != "" {
			if v, ok := rMap[declaredRcpt["id"].(string)]; ok {
				// matched recipient declared by ID
				result[i] = v
				delete(rMap, v.ID)
			}
		} else {
			// group result recipients by type
			for key, rcpt := range rMap {
				if string(rcpt.Type) == declaredRcpt["type"] && rcpt.Target == declaredRcpt["target"] {
					result[i] = rcpt
					delete(rMap, key)
					break
				}
			}
		}
	}

	// append unmatched read recipients to the result
	for _, rcpt := range rMap {
		result = append(result, rcpt)
	}

	return result
}

func expandRecipient(t honeycombio.RecipientType, d *schema.ResourceData) (*honeycombio.Recipient, error) {
	r := &honeycombio.Recipient{
		ID:   d.Id(),
		Type: t,
	}

	switch r.Type {
	case honeycombio.RecipientTypeEmail:
		r.Details.EmailAddress = d.Get("address").(string)
	case honeycombio.RecipientTypePagerDuty:
		r.Details.PDIntegrationKey = d.Get("integration_key").(string)
		r.Details.PDIntegrationName = d.Get("integration_name").(string)
	case honeycombio.RecipientTypeSlack:
		r.Details.SlackChannel = d.Get("channel").(string)
	case honeycombio.RecipientTypeWebhook:
		r.Details.WebhookName = d.Get("name").(string)
		r.Details.WebhookSecret = d.Get("secret").(string)
		r.Details.WebhookURL = d.Get("url").(string)
	default:
		return r, fmt.Errorf("unsupported recipient type %v", r.Type)
	}
	return r, nil
}
