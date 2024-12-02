# Resource: honeycombio_webhook_recipient

`honeycombio_webhook_recipient` allows you to define and manage a Webhook recipient that can be used by Triggers or BurnAlerts notifications.

## Standard Webhook Example

```hcl
resource "honeycombio_webhook_recipient" "prod" {
  name   = "Production Alerts"
  secret = "a63dab148496ecbe04a1a802ca9b95b8"
  url    = "https://my.url.corp.net"
}
```

## Custom Webhook Example

```hcl
resource "honeycombio_webhook_recipient" "prod" {
  name   = "Production Alerts"
  secret = "a63dab148496ecbe04a1a802ca9b95b8"
  url    = "https://my.url.corp.net"
    
  header {
    name = "Authorization"
    value = "Bearer 123"
  }
    
  template {
    type = "trigger"
    body = <<EOT
		{
			"name": " {{ .Name }}",
			"id": " {{ .ID }}",
			"description": " {{ .Description }}",
            "threshold": {
              "op": "{{ .Operator }}",
              "value": "{{ .Threshold }}"
            },
		}
		EOT
  }
    
  variable {
      name          = "severity"
      default_value = "critical"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the Webhook Integration to create.
* `secret` - (Optional) The secret to include when sending the notification to the webhook.
* `url` - (Required) The URL of the endpoint to send the notification to.
* `template` - (Optional) Zero or more configuration blocks (described below) to customize the webhook payload if desired.
* `variable` - (Optional) Zero or more configuration blocks (described below) to define variables to be used in the webhook payload if desired.
* `header` - (Optional) Zero or more configuration blocks (described below) to add custom webhook headers if desired.

When configuring custom webhook payloads, use the `template` block, which accepts the following arguments:

* `type` - (Required) The template type, allowed types are `trigger`, `exhaustion_time`, and `budget_rate`. Only one template block of each type is allowed on a single recipient.
* `body` - (Required) A JSON formatted string to represent the webhook payload.

Optionally, when configuring custom webhooks, use the `variable` block to create custom variables that can be interpolated in a template. 
To configure a variable, at least one `template` block must also be configured.
The `variable` block accepts the following arguments:

* `name` - (Required) The name of the custom variable. Must be an alphanumeric string beginning with a lowercase letter.
* `default_value` - (Optional) The default value for the custom variable, which can be overridden at the alert level.

Optionally, when configuring custom webhooks, use the `header` block to create custom HTTP headers to be included in the webhook request.
Up to five custom headers can be configured. Reserved headers `Content-Type`, `User-Agent`, and `X-Honeycomb-Webhook-Token` cannot be used.
The `header` block accepts the following arguments:

* `name` - (Required) The name or key for the header.
* `value` - (Optional) The value for the header.



## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the recipient.

## Import

Webhook Recipients can be imported by their ID, e.g.

```
$ terraform import honeycombio_webhook_recipient.my_recipient nx2zsegA0dZ
```
