# Resource: honeycombio_email_recipient

`honeycombio_email_recipient` allows you to define and manage an Email recipient that can be used by Triggers or BurnAlerts notifications.

## Example Usage

```hcl
resource "honeycombio_email_recipient" "alerts" {
  address = "alerts@example.com"
}
```

## Argument Reference

The following arguments are supported:

* `address` - (Required) The email address to send the notification to.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the recipient.
