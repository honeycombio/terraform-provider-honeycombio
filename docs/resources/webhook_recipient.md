# Resource: honeycombio_webhook_recipient

`honeycombio_webhook_recipient` allows you to define and manage a Webhook recipient that can be used by Triggers or BurnAlerts notifications.

## Example Usage

```hcl
resource "honeycombio_webhook_recipient" "prod" {
  name   = "Production Alerts"
  secret = "a63dab148496ecbe04a1a802ca9b95b8"
  url    = "https://my.url.corp.net"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the Webhook Integration to create.
* `secret` - (Required) The secret to include when sending the notification to the webhook.
* `url` - (Required) The URL of the endpoint to send the notification to.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the recipient.
