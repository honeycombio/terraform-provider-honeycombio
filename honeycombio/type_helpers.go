package honeycombio

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
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
		rcpt := map[string]interface{}{
			"id":     r.ID,
			"type":   string(r.Type),
			"target": r.Target,
		}
		if r.Details != nil {
			// notification details have been provided
			details := make([]map[string]interface{}, 1)
			details[0] = map[string]interface{}{}
			if r.Details.PDSeverity != "" {
				details[0]["pagerduty_severity"] = string(r.Details.PDSeverity)
			}
			rcpt["notification_details"] = details
		}
		result[i] = rcpt
	}

	return result
}

func expandNotificationRecipients(s []interface{}) []honeycombio.NotificationRecipient {
	recipients := make([]honeycombio.NotificationRecipient, len(s))

	for i, r := range s {
		rMap := r.(map[string]interface{})

		rcpt := honeycombio.NotificationRecipient{
			ID:     rMap["id"].(string),
			Type:   honeycombio.RecipientType(rMap["type"].(string)),
			Target: rMap["target"].(string),
		}
		if v, ok := rMap["notification_details"].([]interface{}); ok && len(v) > 0 {
			// notification details have been provided
			details := v[0].(map[string]interface{})
			if s, ok := details["pagerduty_severity"]; ok {
				rcpt.Details = &honeycombio.NotificationRecipientDetails{
					PDSeverity: honeycombio.PagerDutySeverity(s.(string)),
				}
			}
		}
		recipients[i] = rcpt
	}

	return recipients
}

// Matches read recipients against those declared in HCL and returns
// the Trigger recipients in a stable order grouped by recipient type.
//
// This cannot currently be handled efficiently by a DiffSuppressFunc.
// See: https://github.com/hashicorp/terraform-plugin-sdk/issues/477
func matchNotificationRecipientsWithSchema(readRecipients []honeycombio.NotificationRecipient, declaredRecipients []interface{}) []honeycombio.NotificationRecipient {
	result := []honeycombio.NotificationRecipient{}

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
	for _, declaredRcpt := range declaredRecipients {
		declaredRcpt := declaredRcpt.(map[string]interface{})

		if declaredRcpt["id"] != "" {
			if v, ok := rMap[declaredRcpt["id"].(string)]; ok {
				// matched recipient declared by ID
				result = append(result, v)
				delete(rMap, v.ID)
			}
		} else {
			// group result recipients by type
			for key, rcpt := range rMap {
				if string(rcpt.Type) == declaredRcpt["type"] && rcpt.Target == declaredRcpt["target"] {
					result = append(result, rcpt)
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

func createRecipient(ctx context.Context, d *schema.ResourceData, meta interface{}, t honeycombio.RecipientType) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	r, err := expandRecipient(t, d)
	if err != nil {
		return diag.FromErr(err)
	}

	r, err = client.Recipients.Create(ctx, r)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(r.ID)
	return readRecipient(ctx, d, meta, t)
}

func readRecipient(ctx context.Context, d *schema.ResourceData, meta interface{}, t honeycombio.RecipientType) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	r, err := client.Recipients.Get(ctx, d.Id())
	if err == honeycombio.ErrNotFound {
		d.SetId("")
		return nil
	} else if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(r.ID)
	switch t {
	case honeycombio.RecipientTypeEmail:
		d.Set("address", r.Details.EmailAddress)
	case honeycombio.RecipientTypePagerDuty:
		d.Set("integration_key", r.Details.PDIntegrationKey)
		d.Set("integration_name", r.Details.PDIntegrationName)
	case honeycombio.RecipientTypeSlack:
		d.Set("channel", r.Details.SlackChannel)
	case honeycombio.RecipientTypeWebhook:
		d.Set("name", r.Details.WebhookName)
		d.Set("secret", r.Details.WebhookSecret)
		d.Set("url", r.Details.WebhookURL)
	default:
		return diag.FromErr(fmt.Errorf("unsupported recipient type %v", t))
	}

	return nil
}

func updateRecipient(ctx context.Context, d *schema.ResourceData, meta interface{}, t honeycombio.RecipientType) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	r, err := expandRecipient(t, d)
	if err != nil {
		return diag.FromErr(err)
	}

	r, err = client.Recipients.Update(ctx, r)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(r.ID)
	return readRecipient(ctx, d, meta, t)
}

func deleteRecipient(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	err := client.Recipients.Delete(ctx, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func expandRecipientFilter(f []interface{}) *recipientFilter {
	var value *string
	var valRegexp *regexp.Regexp

	filter := f[0].(map[string]interface{})
	name := filter["name"].(string)
	if v, ok := filter["value"].(string); ok && v != "" {
		value = honeycombio.StringPtr(v)
	}
	if v, ok := filter["value_regex"].(string); ok && v != "" {
		valRegexp = regexp.MustCompile(v)
	}

	switch name {
	case "address":
		return &recipientFilter{Type: honeycombio.RecipientTypeEmail, Value: value, ValueRegex: valRegexp}
	case "channel":
		return &recipientFilter{Type: honeycombio.RecipientTypeSlack, Value: value, ValueRegex: valRegexp}
	case "integration_name":
		return &recipientFilter{Type: honeycombio.RecipientTypePagerDuty, Value: value, ValueRegex: valRegexp}
	case "name", "url":
		return &recipientFilter{Type: honeycombio.RecipientTypeWebhook, Value: value, ValueRegex: valRegexp}
	default:
		return nil
	}

}

// recipientFilter's help match one or more Recipients
type recipientFilter struct {
	Type       honeycombio.RecipientType
	Value      *string
	ValueRegex *regexp.Regexp
}

// IsMatch determine's if a given Recipient matches the filter
func (f *recipientFilter) IsMatch(r honeycombio.Recipient) bool {
	// nil filter fails open
	if f == nil {
		return true
	}
	// types don't match, no point in going further
	if r.Type != f.Type {
		return false
	}

	if f.ValueRegex != nil {
		switch r.Type {
		case honeycombio.RecipientTypeEmail:
			return f.ValueRegex.MatchString(r.Details.EmailAddress)
		case honeycombio.RecipientTypeSlack:
			return f.ValueRegex.MatchString(r.Details.SlackChannel)
		case honeycombio.RecipientTypePagerDuty:
			return f.ValueRegex.MatchString(r.Details.PDIntegrationName)
		case honeycombio.RecipientTypeWebhook:
			return f.ValueRegex.MatchString(r.Details.WebhookName) || f.ValueRegex.MatchString(r.Details.WebhookURL)
		}
	} else if f.Value != nil {
		switch r.Type {
		case honeycombio.RecipientTypeEmail:
			return (r.Details.EmailAddress == *f.Value)
		case honeycombio.RecipientTypeSlack:
			return (r.Details.SlackChannel == *f.Value)
		case honeycombio.RecipientTypePagerDuty:
			return (r.Details.PDIntegrationName == *f.Value)
		case honeycombio.RecipientTypeWebhook:
			return (r.Details.WebhookName == *f.Value) || (r.Details.WebhookURL == *f.Value)
		}
	}

	return true
}
