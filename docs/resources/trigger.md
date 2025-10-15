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

resource "honeycombio_trigger" "example" {
  name        = "Requests are slower than usual"
  description = "Average duration of all requests for the last 10 minutes."

  query_json = data.honeycombio_query_specification.example.json
  dataset    = var.dataset

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

### Trigger with PagerDuty Recipient and Severity

```hcl
variable "dataset" {
  type = string
}

data "honeycombio_recipient" "pd_prod" {
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

resource "honeycombio_trigger" "example" {
  name        = "Requests are slower than usual"
  description = "Average duration of all requests for the last 10 minutes."

  query_json = data.honeycombio_query_specification.example.json
  dataset    = var.dataset

  frequency = 600 // in seconds, 10 minutes

  threshold {
    op             = ">"
    value          = 1000
    exceeded_limit = 3
  }

  recipient {
    id = data.honeycombio_recipient.pd_prod.id

    notification_details {
      pagerduty_severity = "warning"
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

  tags = {
    team = "backend"
    env  = "production"
  }
}
```

### Trigger with Webhook Recipient and Notification Variable

```hcl
variable "dataset" {
    type = string
}

data "honeycombio_recipient" "custom_webhook" {
    type = "webhook"

    detail_filter {
        name  = "name"
        value = "My Custom Webhook"
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

resource "honeycombio_trigger" "example" {
    name        = "Requests are slower than usual"
    description = "Average duration of all requests for the last 10 minutes."

    query_json = data.honeycombio_query_specification.example.json
    dataset    = var.dataset

    frequency = 600 // in seconds, 10 minutes

    threshold {
        op             = ">"
        value          = 1000
        exceeded_limit = 3
    }

    recipient {
        id = data.honeycombio_recipient.custom_webhook.id

        notification_details {
            variable {
                name = "severity"
                value = "info"
            }
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

  tags = {
      team = "backend"
      env  = "production"
  }
}
```

### Baseline Trigger

```hcl
variable "dataset" {
    type = string
}

data "honeycombio_query_specification" "example" {
    calculation {
        op     = "AVG"
        column = "duration_ms"
    }
}

resource "honeycombio_trigger" "example" {
    name        = "Requests are slower than usual"
    description = "Average duration of all requests for the last 10 minutes."

    query_json = data.honeycombio_query_specification.example.json
    dataset    = var.dataset

    frequency = 600 // in seconds, 10 minutes

    threshold {
        op             = ">="
        value          = 1000
    }

    baseline_details {
        type            = "percentage"
        offset_minutes  = 1440
    }

    tags = {
        team = "backend"
        env  = "production"
    }
}
```

### Environment-wide Trigger

```hcl

data "honeycombio_query_specification" "example" {
    calculation {
        op     = "AVG"
        column = "duration_ms"
    }
}

resource "honeycombio_trigger" "example" {
    name        = "Requests are slower than usual"
    description = "Average duration of all requests for the last 10 minutes."

    query_json = data.honeycombio_query_specification.example.json

    frequency = 600 // in seconds, 10 minutes

    threshold {
        op             = ">="
        value          = 1000
    }
}
```

### Trigger with Having

```hcl
variable "dataset" {
  type = string
}

data "honeycombio_query_specification" "query" {
  calculation {
    op     = "AVG"
    column = "duration_ms"
  }

  // `having`'s must have a matching `calculation`. This won't be used as the
  // trigger's threshold since it matches the `having` - `AVG(duration_ms)` will
  // be used for the threshold. To use a `having` to restrict the trigger's
  // threshold, omit the second (different) `calculation`.
  calculation {
    op = "MAX"
    column = "retries"
  }

  filter {
    column = "error.type"
    op = "exists"
  }

  // Only returns results with at least one retry
  having {
    calculate_op = "MAX"
    column = "retries"
    op = ">"
    value = 0
  }

  time_range = 900 // in seconds, 15 minutes
}

resource "honeycombio_trigger" "trigger" {
  name        = "Retried errors are slower than usual"
  description = "Average duration of requests with errors and at least one retry is slower than expected for the last 15 minutes."

  disabled = false

  query_json = data.honeycombio_query_specification.query.json
  dataset    = var.dataset

  frequency = 900 // in seconds, 15 minutes

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
    target = "Trigger - slow requests" # name of the marker
  }
}
```

### Trigger with Having COUNT

```hcl
variable "dataset" {
  type = string
}

data "honeycombio_query_specification" "query" {
  calculation {
    op     = "AVG"
    column = "duration_ms"
  }

  calculation {
    op = "COUNT"
  }

  // Only returns results where more than 100 events were received. Windows with
  // less than 100 events will yield zero, and this trigger will consider them
  // "ok"
  having {
    calculate_op = "COUNT"
    op = ">"
    value = 100
  }

  time_range = 900 // in seconds, 15 minutes
}

resource "honeycombio_trigger" "trigger" {
  name        = "Common requests are slower than usual"
  description = "Average duration of common requests is slower than expected for the last 15 minutes."

  disabled = false

  query_json = data.honeycombio_query_specification.query.json
  dataset    = var.dataset

  frequency = 900 // in seconds, 15 minutes

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
    target = "Trigger - slow requests" # name of the marker
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the trigger.
* `dataset` - (Optional) The dataset this trigger is associated with. If omitted, the lookup will be Environment-wide.
* `query_id` - (Optional) The ID of the Query that the Trigger will execute. Conflicts with `query_json`.
* `query_json` - (Optional) The Query Specfication JSON for the Trigger to execute.
Providing the Query Specification as JSON -- as opposed to a Query ID -- enables additional validation during the validate and plan stages.
Conflicts with `query_id`.
* `threshold` - (Required) A configuration block (described below) describing the threshold of the trigger.
* `description` - (Optional) Description of the trigger.
* `disabled` - (Optional) The state of the trigger. If true, the trigger will not be run. Defaults to false.
* `frequency` - (Optional) The interval (in seconds) in which to check the results of the query’s calculation against the threshold.
This value must be divisible by 60, between 60 and 86400 (between 1 minute and 1 day), and not be more than 4 times the query's duration (see note below).
Defaults to 900 (15 minutes).
* `alert_type` - (Optional) The frequency for the alert to trigger. (`on_change` is the default behavior, `on_true` can also be selected)
* `evaluation_schedule` - (Optional) A configuration block (described below) that determines when the trigger is run.
When the time is within the scheduled window the trigger will be run at the specified frequency.
Outside of the window, the trigger will not be run.
If no schedule is specified, the trigger will be run at the specified frequency at all times.
* `baseline_details` - (Optional) A configuration block (described below) allows you to receive notifications when the delta between values in your data, compared to a previous time period, cross thresholds you configure.
* `recipient` - (Optional) Zero or more configuration blocks (described below) with the recipients to notify when the trigger fires.
* `tags` - (Optional) Map of up to ten (10) tags to assign to the resource.

One of `query_id` or `query_json` are required.

-> **NOTE** The query used in a Trigger must follow a strict subset: the query must contain *exactly one* calcuation and may only contain `calculation`, `filter`, `filter_combination` and `breakdowns` fields.
The query's duration cannot be more than four times the trigger frequency (i.e. `duration <= frequency*4`).
See [A Caveat on Time](https://docs.honeycomb.io/working-with-your-data/query-specification/#a-caveat-on-time)) for more information on specifying a query's duration.
For example: if using the default query `time_range` of `7200` the lowest `frequency` for a trigger is `1800`.

Each trigger configuration must contain exactly one `threshold` block, which accepts the following arguments:

* `op` - (Required) The operator to apply, allowed threshold operators are `>`, `>=`, `<`, and `<=`.
* `value` - (Required) The value to be used with the operator.
* `exceeded_limit` - (Optional) The number of times the threshold is met before an alert is sent, must be between 1 and 5. Defaults to `1`.

Each trigger configuration may provide an `evaluation_schedule` block, which accepts the following arguments:

* `start_time` - (Required) UTC time to start evaluating the trigger in HH:mm format (e.g. `13:00`)
* `end_time` - (Required) UTC time to stop evaluating the trigger in HH:mm format (e.g. `13:00`)
* `days_of_week` - (Required) A list of days of the week (in lowercase) to evaluate the trigger on

Each trigger configuration may provide an `baseline_details` block, which accepts the following arguments:

* `type` - (Required) Either `value` or `percentage` to indicate either an absolute value or percentage delta.
* `offset_minutes` - (Required) Either `60` (1 hour), `1440` (24 hours), `10080` (7 days), or `40320` (28 days) to indicate the length of the previous time period.

Each trigger configuration may have zero or more `recipient` blocks, which each accept the following arguments. A trigger recipient block can either refer to an existing recipient (a recipient that is already present in another trigger) or a new recipient. When specifying an existing recipient, only `id` may be set. If you pass in a recipient without its ID and only include the type and target, Honeycomb will make a best effort to match to an existing recipient. To retrieve the ID of an existing recipient, refer to the [`honeycombio_recipient`](../data-sources/recipient.md) data source.

* `type` - (Optional) The type of the trigger recipient, allowed types are `email`, `marker`, `msteams`, `pagerduty`, `slack` and `webhook`.
Cannot not be used in combination with `id`.
* `target` - (Optional) Target of the trigger recipient, this has another meaning depending on the type of recipient (see the table below).
Cannot not be used in combination with `id`.
* `id` - (Optional) The ID of an already existing recipient. Cannot not be used in combination with `type` and `target`.
* `notification_details` - (Optional) a block of additional details to send along with the notification.
  * `pagerduty_severity` - (Optional) Indicates the severity of an alert and has a default value of `critical` but can be set to one of `info`, `warning`, `error`, or `critical` and must be used in combination with a PagerDuty recipient.
  * `variable` - (Optional) Up to 10 configuration blocks with a `name` and a `value` to override the default variable value. Must be used in combination with a Webhook recipient that already has a variable with the same name configured.

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

Single-dataset Triggers can be imported using by using their ID combined with their dataset.
If the Trigger is an environment-wide trigger, the dataset is not provided.

### Trigger

```
$ terraform import honeycombio_trigger.my_trigger my-dataset/bj9BwOb1uKz
```

### Environment-wide Trigger

```
$ terraform import honeycombio_trigger.my_trigger bj9BwOb1uJz
```

You can find the ID in the URL bar when visiting the trigger from the UI.
