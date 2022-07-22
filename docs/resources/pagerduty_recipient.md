# Resource: honeycombio_pagerduty_recipient

`honeycombio_pagerduty_recipient` allows you to define and manage a PagerDuty recipient that can be used by Triggers or BurnAlerts notifications.

## Example Usage

```hcl
resource "honeycombio_pagerduty_recipient" "prod-oncall" {
  integration_key  = "cd6e8de3c857aefc950e0d5ebcb79ac2"
  integration_name = "Production on-call notifications"
}
```

## Argument Reference

The following arguments are supported:

* `integration_key` - (Required) The key of the PagerDuty Integration to send the notification to.
* `integration_name` - (Required) The name of the PagerDuty Integration to send the notification to.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the recipient.

## Import

PagerDuty Recipients can be imported by their ID, e.g.

```
$ terraform import honeycombio_pagerduty_recipient.my_recipient nx2zsegA0dZ
```
