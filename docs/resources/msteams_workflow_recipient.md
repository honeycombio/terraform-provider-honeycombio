# Resource: honeycombio_msteams_workflow_recipient

`honeycombio_msteams_workflow_recipient` allows you to define and manage an MSTeams Workflows recipient that can be used by Triggers or BurnAlerts notifications.

## Example Usage

```hcl
resource "honeycombio_msteams_workflow_recipient" "prod" {
  name = "Production Alerts"
  url  = "https://mycorp.westus.logic.azure.com/workflows/123456"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the recipient.
* `url` - (Required) The MSTeams Workflow URL to send the notification to.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the recipient.

## Import

MSTeams Workflow Recipients can be imported by their ID, e.g.

```
$ terraform import honeycombio_msteams_workflow_recipient.my_recipient nx2zsefB1cX
```
