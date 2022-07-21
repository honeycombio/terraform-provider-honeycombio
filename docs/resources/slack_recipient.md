# Resource: honeycombio_slack_recipient

`honeycombio_slack_recipient` allows you to define and manage a Slack channel or user recipient that can be used by Triggers or BurnAlerts notifications.

## Example Usage

```hcl
resource "honeycombio_slack_recipient" "alerts" {
  channel = "#alerts"
}
```

## Argument Reference

The following arguments are supported:

* `channel` - (Required) The Slack channel or username to send the notification to. Must begin with `#` or `@`.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the recipient.
