# Resource: honeycombio_trigger

Creates a trigger. For more information about triggers, check out [Alert with Triggers](https://docs.honeycomb.io/working-with-your-data/triggers/).

## Example Usage

### Basic Example

```hcl
variable "dataset" {
  type = string
}

data "honeycombio_query_specification" "example" {
  calculation {
    op     = "AVG"
    column = "duration_ms"
  }

  filter {
    column = "trace.parent_id"
    op     = "does-not-exist"
  }

  time_range = 1800
}

resource "honeycombio_query" "example" {
  dataset    = var.dataset
  query_json = data.honeycombio_query_specification.example.json
}

resource "honeycombio_trigger" "example" {
  name        = "Requests are slower than usual"
  description = "Average duration of all requests for the last 10 minutes."

  query_id = honeycombio_query.example.id
  dataset  = var.dataset

  frequency = 600 // in seconds, 10 minutes

  alert_type = "on_change" // on_change is default, on_true can refers to the "Alert on True" checkbox in the UI

  threshold {
    op    = ">"
    value = 1000
  }

  # zero or more recipients
  recipient {
    type   = "email"
    target = "hello@example.com"
  }

  recipient {
    type   = "marker"
    target = "Trigger - requests are slow"
  }
}
```

### Example with PagerDuty Recipient and Severity
```
variable "dataset" {
  type = string
}

data "honeycombio_recipient" "pd-prod" {
  type = "pagerduty"

  detail_filter {
    name  = "integration_name"
    value = "Prod On-Call"
  }
}

data "honeycombio_query_specification" "example" {
  calculation {
    op     = "AVG"
    column = "duration_ms"
  }

  filter {
    column = "trace.parent_id"
    op     = "does-not-exist"
  }
}

resource "honeycombio_query" "example" {
  dataset    = var.dataset
  query_json = data.honeycombio_query_specification.example.json
}

resource "honeycombio_trigger" "example" {
  name        = "Requests are slower than usual"
  description = "Average duration of all requests for the last 10 minutes."

  query_id = honeycombio_query.example.id
  dataset  = var.dataset

  frequency = 600 // in seconds, 10 minutes

  threshold {
    op    = ">"
    value = 1000
  }

  recipient {
    id = data.honeycombio_recipient.pd-prod.id

    notification_details {
      pagerduty_severity = "info"
    }
  }

  evaluation_schedule {
    start_time = "13:00"
    end_time   = "21:00"

    days_of_week = [
      "monday",
      "wednesday",
      "friday"
    ]
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the trigger.
* `dataset` - (Required) The dataset this trigger is associated with.
* `query_id` - (Required) The ID of the Query that the Trigger will execute.
* `threshold` - (Required) A configuration block (described below) describing the threshold of the trigger.
* `description` - (Optional) Description of the trigger.
* `disabled` - (Optional) The state of the trigger. If true, the trigger will not be run. Defaults to false.
* `frequency` - (Optional) The interval (in seconds) in which to check the results of the query’s calculation against the threshold.
This value must be divisible by 60, between 60 and 86400 (between 1 minute and 1 day), and not be more than 4 times the query's duration.
Defaults to 900 (15 minutes).
* `alert_type` - (Optional) The frequency for the alert to trigger. (`on_change` is the default behavior, `on_true` can also be selected)
* `evaluation_schedule` - (Optional) A configuration block (described below) that determines when the trigger is run.
When the time is within the scheduled window the trigger will be run at the specified frequency.
Outside of the window, the trigger will not be run.
If no schedule is specified, the trigger will be run at the specified frequency at all times.
* `recipient` - (Optional) Zero or more configuration blocks (described below) with the recipients to notify when the trigger fires.

-> **NOTE** The query used in a Trigger must follow a strict subset: the query must contain *exactly one* calcuation and may only contain `calculation`, `filter`, `filter_combination` and `breakdowns` fields.
The query's duration (`time_range` in the specification) cannot be more than four times the trigger frequency. For example: if using the default query `time_range` of `7200` the lowest `frequency` for a trigger is `1800`.

Each trigger configuration must contain exactly one `threshold` block, which accepts the following arguments:

* `op` - (Required) The operator to apply, allowed threshold operators are `>`, `>=`, `<`, and `<=`.
* `value` - (Required) The value to be used with the operator.

Each trigger configuration may provide an `evaluation_schedule` block, which accepts the following arguments:

* `start_time` - (Required) UTC time to start evaluating the trigger in HH:mm format (e.g. `13:00`)
* `end_time` - (Required) UTC time to start evaluating the trigger in HH:mm format (e.g. `13:00`)
* `days_of_week` - (Required) A list of days of the week (in lowercase) to evaluate the trigger on

Each trigger configuration may have zero or more `recipient` blocks, which each accept the following arguments. A trigger recipient block can either refer to an existing recipient (a recipient that is already present in another trigger) or a new recipient. When specifying an existing recipient, only `id` may be set. If you pass in a recipient without its ID and only include the type and target, Honeycomb will make a best effort to match to an existing recipient. To retrieve the ID of an existing recipient, refer to the [`honeycombio_recipient`](../data-sources/recipient.md) data source.

* `type` - (Optional) The type of the trigger recipient, allowed types are `email`, `marker`, `pagerduty`, `slack` and `webhook`.
Cannot not be used in combination with `id`.
* `target` - (Optional) Target of the trigger recipient, this has another meaning depending on the type of recipient (see the table below).
Cannot not be used in combination with `id`.
* `id` - (Optional) The ID of an already existing recipient. Cannot not be used in combination with `type` and `target`.
* `notification_details` - (Optional) a block of additional details to send along with the notification. The only supported option currently is `pagerduty_severity` which has a default value of `critical` but can be set to one of `info`, `warning`, `error`, or `critical` and must be used in combination with a PagerDuty recipient.

Type      | Target
----------|-------------------------
email     | an email address
marker    | name of the marker
pagerduty | _N/A_
slack     | name of the channel
webhook   | name of the webhook

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the trigger.

## Import

Triggers can be imported using a combination of the dataset name and their ID, e.g.

```
$ terraform import honeycombio_trigger.my_trigger my-dataset/AeZzSoWws9G
```

You can find the ID in the URL bar when visiting the trigger from the UI.
