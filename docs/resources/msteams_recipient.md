# Resource: honeycombio_msteams_recipient

`honeycombio_msteams_recipient` allows you to define and manage an MSTeams recipient that can be used by Triggers or BurnAlerts notifications.

## Example Usage

```hcl
resource "honeycombio_msteams_recipient" "prod" {
  name = "Production Alerts"
  url  = "https://mycorp.webhook.office.com/webhookb2/abcd12345"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the MS Teams Integration to create.
* `url` - (Required) The Incoming Webhook URL to send the notification to.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the recipient.

## Import

MSTeams Recipients can be imported by their ID, e.g.

```
$ terraform import honeycombio_msteams_recipient.my_recipient nx2zsefB1cX
```
