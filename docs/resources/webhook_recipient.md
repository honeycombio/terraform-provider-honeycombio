# Resource: honeycombio_webhook_recipient

`honeycombio_webhook_recipient` allows you to define and manage a Webhook recipient that can be used by Triggers or BurnAlerts notifications.

## Example Usage

```hcl
resource "honeycombio_webhook_recipient" "prod" {
  name   = "Production Alerts"
  secret = "a63dab148496ecbe04a1a802ca9b95b8"
  url    = "https://my.url.corp.net"

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
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the Webhook Integration to create.
* `secret` - (Optional) The secret to include when sending the notification to the webhook.
* `url` - (Required) The URL of the endpoint to send the notification to.
* `template` - (Optional) Zero or more configuration blocks (described below) to customize the webhook payload if desired.

When configuring custom webhook payloads, use the `template` block, which accepts the following arguments:

* `type` - (Required) The template type, allowed types are `trigger`, `exhaustion_time`, and `budget_rate`. Only one template block of each type is allowed on a single recipient.
* `body` - (Required) A JSON formatted string to represent the webhook payload.


## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the recipient.

## Import

Webhook Recipients can be imported by their ID, e.g.

```
$ terraform import honeycombio_webhook_recipient.my_recipient nx2zsegA0dZ
```
