# Resource: honeycombio_trigger

Creates a trigger. For more information about triggers, check out [Alert with Triggers](https://docs.honeycomb.io/working-with-your-data/triggers/).

## Example Usage

### Basic Example

```terraform
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

```terraform
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

```terraform
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

```terraform
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

```terraform
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

```terraform
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

```terraform
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

### Trigger with Formula

```terraform
variable "dataset" {
  type = string
}

data "honeycombio_query_specification" "query" {
  calculation {
    op   = "COUNT"
    name = "total"
  }

  calculation {
    op   = "COUNT"
    name = "errors"

    filter {
      column = "error"
      op     = "exists"
    }
  }

  formula {
    name       = "error_rate"
    expression = "DIV($errors, $total)"
  }

  time_range = 900 // in seconds, 15 minutes
}

resource "honeycombio_trigger" "trigger" {
  name        = "Error rate is too high"
  description = "The error rate has exceeded the threshold for the last 15 minutes."

  query_json = data.honeycombio_query_specification.query.json
  dataset    = var.dataset

  frequency = 900 // in seconds, 15 minutes

  threshold {
    op    = ">"
    value = 0.1
  }

  recipient {
    type   = "email"
    target = "hello@example.com"
  }
}
```

### Metrics Trigger (Simple)

```terraform
variable "dataset" {
  type = string
}

data "honeycombio_query_specification" "metrics" {
  calculation {
    op     = "P95"
    column = "http.server.request.duration"
  }

  time_range  = 1800
  granularity = 300 # Custom granularity only available with Metrics
}

resource "honeycombio_trigger" "metrics" {
  name    = "High request duration"
  dataset = var.dataset

  query_json = data.honeycombio_query_specification.metrics.json

  frequency = 900

  threshold {
    op    = ">"
    value = 1000
  }

  recipient {
    type   = "email"
    target = "hello@example.com"
  }
}
```

### Metrics Trigger (Custom Temporal Aggregation)

```terraform
variable "dataset" {
  type = string
}

data "honeycombio_query_specification" "metrics" {
  calculated_field {
    name       = "request_rate_5m"
    expression = "INCREASE($http.server.requests, 300)" # 5-minute range interval
  }

  calculation {
    op     = "AVG"
    column = "request_rate_5m"
  }

  time_range  = 1800
  granularity = 60 # 1-minute time step, but rate is calculated over 5 minutes
}

resource "honeycombio_trigger" "metrics" {
  name    = "High request rate (custom temporal aggregation)"
  dataset = var.dataset

  query_json = data.honeycombio_query_specification.metrics.json

  frequency = 900

  threshold {
    op    = ">"
    value = 1000
  }

  recipient {
    type   = "email"
    target = "hello@example.com"
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) The name of the Trigger.

### Optional

- `alert_type` (String) Control when the Trigger will send a notification.
- `baseline_details` (Block List) A configuration block that allows you to receive notifications when the delta between values in your data, compared to a previous time period, cross thresholds you configure. (see [below for nested schema](#nestedblock--baseline_details))
- `dataset` (String) The dataset this Trigger is associated with.
- `description` (String) A description of the Trigger.
- `disabled` (Boolean) The state of the Trigger. If true, the Trigger will not be run.
- `evaluation_schedule` (Block List) The schedule that determines when the trigger is run. When the time is within the scheduled window,  the trigger will be run at the specified frequency. Outside of the window, the trigger will not be run.If no schedule is specified, the trigger will be run at the specified frequency at all times. (see [below for nested schema](#nestedblock--evaluation_schedule))
- `frequency` (Number) The interval (in seconds) in which to check the results of the query's calculation against the threshold. This value must be divisible by 60, between 60 and 86400 (between 1 minute and 1 day), and not be more than 4 times the query's duration.
- `query_id` (String) The ID of the Query that the Trigger will execute.
- `query_json` (String) The QuerySpec JSON for the query that the Trigger will execute. Providing the QuerySpec JSON directly allows for additional validation that the QuerySpec is valid as a Trigger Query. While the JSON can be constructed manually, it is easiest to use the `honeycombio_query_specification` data source.
- `recipient` (Block Set) Zero or more recipients to notify when the resource fires. (see [below for nested schema](#nestedblock--recipient))
- `tags` (Map of String) A map of tags to assign to the resource.
- `threshold` (Block List) A block describing the threshold for the Trigger to fire. (see [below for nested schema](#nestedblock--threshold))

### Read-Only

- `id` (String) The unique identifier for this Trigger.

<a id="nestedblock--baseline_details"></a>
### Nested Schema for `baseline_details`

Required:

- `offset_minutes` (Number) What previous time period to evaluate against: 1 hour, 1 day, 1 week, or 4 weeks.
- `type` (String) Whether to use an absolute value or percentage delta.


<a id="nestedblock--evaluation_schedule"></a>
### Nested Schema for `evaluation_schedule`

Required:

- `days_of_week` (List of String) The days of the week to evaluate the trigger on
- `end_time` (String) UTC time to stop evaluating the trigger in HH:mm format
- `start_time` (String) UTC time to start evaluating the trigger in HH:mm format


<a id="nestedblock--recipient"></a>
### Nested Schema for `recipient`

Optional:

- `id` (String) The ID of an existing recipient.
- `notification_details` (Block List) Additional details to send along with the notification. (see [below for nested schema](#nestedblock--recipient--notification_details))
- `target` (String) Target of the notification, this has another meaning depending on the type of recipient.
- `type` (String) The type of the notification recipient.

<a id="nestedblock--recipient--notification_details"></a>
### Nested Schema for `recipient.notification_details`

Optional:

- `pagerduty_severity` (String) The severity to set with the PagerDuty notification. If no severity is provided, 'critical' is assumed.
- `variable` (Block Set) The variables to set with the webhook notification. (see [below for nested schema](#nestedblock--recipient--notification_details--variable))

<a id="nestedblock--recipient--notification_details--variable"></a>
### Nested Schema for `recipient.notification_details.variable`

Required:

- `name` (String) The name of the variable

Optional:

- `value` (String) The value of the variable




<a id="nestedblock--threshold"></a>
### Nested Schema for `threshold`

Required:

- `op` (String) The operator to apply.
- `value` (Number) The value to be used with the operator.

Optional:

- `exceeded_limit` (Number) The number of times the threshold is met before an alert is sent. Defaults to 1.

-> **NOTE** The query used in a Trigger must follow a strict subset. It supports two query shapes:
**Standard:** The query must contain *exactly one* non-having calculation (without names or aggregate filters) and may only contain `calculation`, `filter`, `filter_combination`, `having` (at most 1), and `breakdowns` fields.
**Formula:** The query must contain *exactly one* formula with up to 100 named calculations. When calculations use names or aggregate-level filters, global filters cannot be used — use calculation-level filters instead.
The query's duration cannot be more than four times the trigger frequency (i.e. `duration <= frequency*4`).
See [A Caveat on Time](https://docs.honeycomb.io/working-with-your-data/query-specification/#a-caveat-on-time)) for more information on specifying a query's duration.
For example: if using the default query `time_range` of `7200` the lowest `frequency` for a trigger is `1800`.

Each `recipient` block can either refer to an existing recipient (already present in another burn alert or trigger) or a new recipient.
When specifying an existing recipient, only `id` may be set.
To retrieve the ID of an existing recipient, refer to the [`honeycombio_recipient`](../data-sources/recipient.md) data source.

| Type      | Target              |
| --------- | ------------------- |
| email     | an email address    |
| marker    | name of the marker  |
| pagerduty | _N/A_               |
| slack     | name of the channel |
| webhook   | name of the webhook |

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
