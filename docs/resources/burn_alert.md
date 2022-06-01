# Resource: honeycombio_burn_alert

Creates a burn alert. For more information about burn alerts, check out [Define Burn Alerts](https://docs.honeycomb.io/working-with-your-data/slos/slo-process/#define-burn-alerts).

## Example Usage

```hcl
variable "dataset" {
  type = string
}

variable "slo_id" {
  type = string
}

resource "honeycombio_burn_alert" "example_alert" {
  dataset            = var.dataset
  slo_id             = var.slo_id
  exhaustion_minutes = 480

  # zero or more recipients
  recipient {
    type   = "email"
    target = "hello@example.com"
  }

  recipient {
    type   = "slack"
    target = "#example-channel"
  }
}
```

## Argument Reference

The following arguments are supported:

* `slo_id` - (Required) ID of the SLO this burn alert is associated with.
* `dataset` - (Required) The dataset this burn alert is associated with.
* `exhaustion_minutes` - (Required) The amount of time, in minutes, remaining before the SLO's error budget will be exhausted and the alert will fire.
* `recipient` - (Optional) Zero or more configuration blocks (described below) with the recipients to notify when the alert fires.

Each burn alert configuration may have zero or more `recipient` blocks, which each accept the following arguments. A recipient block can either refer to an existing recipient (a recipient that is already present in another burn alert or trigger) or a new recipient. When specifying an existing recipient, only `id` may be set. If you pass in a recipient without its ID and only include the type and target, Honeycomb will make a best effort to match to an existing recipient. To retrieve the ID of an existing recipient, refer to the [`honeycombio_recipient`](../data-sources/recipient.md) data source.

* `type` - (Optional) The type of the recipient, allowed types are `email`, `pagerduty`, `slack` and `webhook`. Should not be used in combination with `id`.
* `target` - (Optional) Target of the recipient, this has another meaning depending on the type of recipient (see the table below). Should not be used in combination with `id`.
* `id` - (Optional) The ID of an already existing recipient. Should not be used in combination with `type` and `target`.

Type      | Target
----------|-------------------------
email     | an email address
pagerduty | _N/A_
slack     | name of the channel
webhook   | name of the webhook

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the burn alert.
