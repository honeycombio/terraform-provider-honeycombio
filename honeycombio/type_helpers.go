package honeycombio

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
)

// getDatasetOrAll returns the dataset from the resource data.
// If the dataset is empty, it returns the 'magic' EnvironmentWideSlug `__all__`.
func getDatasetOrAll(d *schema.ResourceData) string {
	dataset, ok := d.Get("dataset").(string)
	if !ok {
		// This is a helper function, so we'll return the default if type assertion fails
		return honeycombio.EnvironmentWideSlug
	}
	if dataset == "" {
		dataset = honeycombio.EnvironmentWideSlug
	}
	return dataset
}

func expandRecipient(t honeycombio.RecipientType, d *schema.ResourceData) (*honeycombio.Recipient, error) {
	r := &honeycombio.Recipient{
		ID:   d.Id(),
		Type: t,
	}

	switch r.Type {
	case honeycombio.RecipientTypeEmail:
		address, ok := d.Get("address").(string)
		if !ok {
			return nil, fmt.Errorf("address must be a string")
		}
		r.Details.EmailAddress = address
	case honeycombio.RecipientTypePagerDuty:
		integrationKey, ok := d.Get("integration_key").(string)
		if !ok {
			return nil, fmt.Errorf("integration_key must be a string")
		}
		integrationName, ok := d.Get("integration_name").(string)
		if !ok {
			return nil, fmt.Errorf("integration_name must be a string")
		}
		r.Details.PDIntegrationKey = integrationKey
		r.Details.PDIntegrationName = integrationName
	case honeycombio.RecipientTypeSlack:
		channel, ok := d.Get("channel").(string)
		if !ok {
			return nil, fmt.Errorf("channel must be a string")
		}
		r.Details.SlackChannel = channel
	case honeycombio.RecipientTypeMSTeams, honeycombio.RecipientTypeMSTeamsWorkflow: //nolint:staticcheck
		name, ok := d.Get("name").(string)
		if !ok {
			return nil, fmt.Errorf("name must be a string")
		}
		url, ok := d.Get("url").(string)
		if !ok {
			return nil, fmt.Errorf("url must be a string")
		}
		r.Details.WebhookName = name
		r.Details.WebhookURL = url
	case honeycombio.RecipientTypeWebhook:
		webhookName, ok := d.Get("name").(string)
		if !ok {
			return nil, fmt.Errorf("name must be a string")
		}
		secret, ok := d.Get("secret").(string)
		if !ok {
			return nil, fmt.Errorf("secret must be a string")
		}
		url, ok := d.Get("url").(string)
		if !ok {
			return nil, fmt.Errorf("url must be a string")
		}
		r.Details.WebhookName = webhookName
		r.Details.WebhookSecret = secret
		r.Details.WebhookURL = url
	default:
		return r, fmt.Errorf("unsupported recipient type %v", r.Type)
	}
	return r, nil
}

func createRecipient(ctx context.Context, d *schema.ResourceData, meta interface{}, t honeycombio.RecipientType) diag.Diagnostics {
	client, err := getConfiguredClient(meta)
	if err != nil {
		return diagFromErr(err)
	}

	r, err := expandRecipient(t, d)
	if err != nil {
		return diag.FromErr(err)
	}

	r, err = client.Recipients.Create(ctx, r)
	if err != nil {
		return diagFromErr(err)
	}

	d.SetId(r.ID)
	return readRecipient(ctx, d, meta, t)
}

func readRecipient(ctx context.Context, d *schema.ResourceData, meta interface{}, t honeycombio.RecipientType) diag.Diagnostics {
	client, err := getConfiguredClient(meta)
	if err != nil {
		return diagFromErr(err)
	}

	var detailedErr honeycombio.DetailedError
	r, err := client.Recipients.Get(ctx, d.Id())
	if errors.As(err, &detailedErr) {
		if detailedErr.IsNotFound() {
			d.SetId("")
			return nil
		} else {
			return diagFromDetailedErr(detailedErr)
		}
	} else if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(r.ID)
	switch t {
	case honeycombio.RecipientTypeEmail:
		_ = d.Set("address", r.Details.EmailAddress)
	case honeycombio.RecipientTypePagerDuty:
		_ = d.Set("integration_key", r.Details.PDIntegrationKey)
		_ = d.Set("integration_name", r.Details.PDIntegrationName)
	case honeycombio.RecipientTypeSlack:
		_ = d.Set("channel", r.Details.SlackChannel)
	case honeycombio.RecipientTypeMSTeams, honeycombio.RecipientTypeMSTeamsWorkflow: //nolint:staticcheck
		_ = d.Set("name", r.Details.WebhookName)
		_ = d.Set("url", r.Details.WebhookURL)
	case honeycombio.RecipientTypeWebhook:
		_ = d.Set("name", r.Details.WebhookName)
		_ = d.Set("secret", r.Details.WebhookSecret)
		_ = d.Set("url", r.Details.WebhookURL)
	default:
		return diag.FromErr(fmt.Errorf("unsupported recipient type %v", t))
	}

	return nil
}

func updateRecipient(ctx context.Context, d *schema.ResourceData, meta interface{}, t honeycombio.RecipientType) diag.Diagnostics {
	client, err := getConfiguredClient(meta)
	if err != nil {
		return diagFromErr(err)
	}

	r, err := expandRecipient(t, d)
	if err != nil {
		return diag.FromErr(err)
	}

	r, err = client.Recipients.Update(ctx, r)
	if err != nil {
		return diagFromErr(err)
	}

	d.SetId(r.ID)
	return readRecipient(ctx, d, meta, t)
}

func deleteRecipient(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := getConfiguredClient(meta)
	if err != nil {
		return diagFromErr(err)
	}

	err = client.Recipients.Delete(ctx, d.Id())
	if err != nil {
		return diagFromErr(err)
	}
	return nil
}

func expandRecipientFilter(f []interface{}) *recipientFilter {
	var value *string
	var valRegexp *regexp.Regexp

	filter, ok := f[0].(map[string]interface{})
	if !ok {
		return nil
	}
	name, ok := filter["name"].(string)
	if !ok {
		return nil
	}
	if v, ok := filter["value"].(string); ok && v != "" {
		value = honeycombio.ToPtr(v)
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

	if f.ValueRegex != nil {
		switch r.Type {
		case honeycombio.RecipientTypeEmail:
			return f.ValueRegex.MatchString(r.Details.EmailAddress)
		case honeycombio.RecipientTypeSlack:
			return f.ValueRegex.MatchString(r.Details.SlackChannel)
		case honeycombio.RecipientTypePagerDuty:
			return f.ValueRegex.MatchString(r.Details.PDIntegrationName)
		case honeycombio.RecipientTypeWebhook,
			honeycombio.RecipientTypeMSTeams, //nolint:staticcheck
			honeycombio.RecipientTypeMSTeamsWorkflow:
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
		case honeycombio.RecipientTypeWebhook,
			honeycombio.RecipientTypeMSTeams, //nolint:staticcheck
			honeycombio.RecipientTypeMSTeamsWorkflow:
			return (r.Details.WebhookName == *f.Value) || (r.Details.WebhookURL == *f.Value)
		}
	}

	return true
}

// diagFromErr is a helper function that converts an error to a diag.Diagnostics.
// Intended to be a drop-in replacement for diag.FromErr from the V2 Plugin SDK.
//
// If err is a honeycombio.DetailedError, a detailed Diagnostic will be added to diag,
// otherwise a generic error Diagnostic will be added to diag.
func diagFromErr(err error) diag.Diagnostics {
	if err == nil {
		return nil
	}

	var detailedErr honeycombio.DetailedError
	if errors.As(err, &detailedErr) {
		return diagFromDetailedErr(detailedErr)
	}

	return diag.FromErr(err)
}

func diagFromDetailedErr(err honeycombio.DetailedError) diag.Diagnostics {
	diags := make(diag.Diagnostics, 0, len(err.Details)+1)
	if len(err.Details) > 0 {
		for _, d := range err.Details {
			detail := d.Code + " - "
			if d.Field != "" {
				detail += d.Field + " "
			}
			detail += d.Description

			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  err.Title,
				Detail:   detail,
			})
		}
	} else {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  err.Message,
			Detail:   err.Title,
		})
	}

	return diags
}
