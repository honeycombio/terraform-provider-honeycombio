# Resource: honeycombio_burn_alert

Creates a burn alert. 

For more information about burn alerts, 
check out [Define Burn Alerts](https://docs.honeycomb.io/working-with-your-data/slos/burn-alerts).

## Example Usage

### Basic Example - Exhaustion Time Burn Alert
```hcl
variable "dataset" {
  type = string
}

variable "slo_id" {
  type = string
}

resource "honeycombio_burn_alert" "example_alert" {
  alert_type         = "exhaustion_time"
  exhaustion_minutes = 480

  dataset = var.dataset
  slo_id  = var.slo_id

  # one or more recipients
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

### Basic Example - Budget Rate Burn Alert
```hcl
variable "dataset" {
  type = string
}

variable "slo_id" {
  type = string
}

resource "honeycombio_burn_alert" "example_alert" {
  alert_type                   = "budget_rate"
  budget_rate_window_minutes   = 480
  budget_rate_decrease_percent = 1

  dataset = var.dataset
  slo_id  = var.slo_id

  # one or more recipients
  recipient {
    type   = "webhook"
    target = "name of the webhook"
  }
}
```

### Example - Exhaustion Time Burn Alert with PagerDuty Recipient and Severity
```hcl
variable "dataset" {
  type = string
}

variable "slo_id" {
  type = string
}

data "honeycombio_recipient" "pd_prod" {
  type = "pagerduty"

  detail_filter {
    name  = "integration_name"
    value = "Prod On-Call"
  }
}

resource "honeycombio_burn_alert" "example_alert" {
  exhaustion_minutes = 60
  dataset            = var.dataset
  slo_id             = var.slo_id

  recipient {
    id = data.honeycombio_recipient.pd_prod.id

    notification_details {
      pagerduty_severity = "critical"
    }
  }
}
```


## Argument Reference

The following arguments are supported:
* `slo_id` - (Required) ID of the SLO this burn alert is associated with.
* `dataset` - (Required) The dataset this burn alert is associated with.
* `alert_type` - (Optional) Type of the burn alert. Valid values are `exhaustion_time` and `budget_rate`. 
   Defaults to `exhaustion_time`.
* `budget_rate_window_minutes` - (Optional) The time period, in minutes, over which a budget rate will be calculated. 
   Must be between 60 and the associated SLO's time period.
   Required when `alert_type` is `budget_rate`.
   Must not be provided when `alert_type` is `exhaustion_time`.
* `budget_rate_decrease_percent` - (Optional) The percent the budget has decreased over the budget rate window.
   The alert will fire when this budget decrease threshold is reached.
   Must be between 0.0001% and 100%, with no more than 4 numbers past the decimal point.
   Required when `alert_type` is `budget_rate`.
   Must not be provided when `alert_type` is `exhaustion_time`.
* `exhaustion_minutes` - (Optional) The amount of time, in minutes, remaining before the SLO's error budget will be exhausted and 
   the alert will fire.
   Must be 0 or greater.
   Required when `alert_type` is `exhaustion_time`.
   Must not be provided when `alert_type` is `budget_rate`.
* `recipient` - (Required) Zero or more configuration blocks (described below) with the recipients to notify when the alert fires.

Each burn alert configuration may have one or more `recipient` blocks, which each accept the following arguments. A recipient block can either refer to an existing recipient (a recipient that is already present in another burn alert or trigger) or a new recipient. When specifying an existing recipient, only `id` may be set. If you pass in a recipient without its ID and only include the type and target, Honeycomb will make a best effort to match to an existing recipient. To retrieve the ID of an existing recipient, refer to the [`honeycombio_recipient`](../data-sources/recipient.md) data source.

* `type` - (Optional) The type of the recipient, allowed types are `email`, `pagerduty`, `msteams`, `slack` and `webhook`. Should not be used in combination with `id`.
* `target` - (Optional) Target of the recipient, this has another meaning depending on the type of recipient (see the table below). Should not be used in combination with `id`.
* `id` - (Optional) The ID of an already existing recipient. Should not be used in combination with `type` and `target`.
* `notification_details` - (Optional) a block of additional details to send along with the notification. The only supported option currently is `pagerduty_severity` which has a default value of `critical` but can be set to one of `info`, `warning`, `error`, or `critical` and must be used in combination with a PagerDuty recipient.

| Type      | Target              |
|-----------|---------------------|
| email     | an email address    |
| pagerduty | _N/A_               |
| slack     | name of the channel |
| webhook   | name of the webhook |

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the burn alert.

## Import

Burn alerts can be imported using a combination of the dataset name and their ID, e.g.

```
$ terraform import honeycombio_burn_alert.my_alert my-dataset/bj9BwOb1uKz
```
