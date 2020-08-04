# Resource: honeycombio_trigger

Creates a trigger. For more information about triggers, check out [Alert with Triggers](https://docs.honeycomb.io/working-with-your-data/triggers/).

## Example Usage

```hcl
variable "dataset" {
    type = string
}

data "honeycombio_query" "example" {
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
  name        = "Requests are slower than usuals"
  description = "Average duration of all requests for the last 10 minutes."
  dataset     = var.dataset

  query_json = data.honeycombio_query.example.rendered

  frequency = 600 // in seconds, 10 minutes

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
    type   = "email"
    target = "bye@example.com"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the trigger.
* `description` - (Optional) Description of the trigger.
* `dataset` - (Required) The dataset this trigger is associated with.
* `disabled` - (Optional) The state of the trigger. If true, the trigger will not be run. Defaults to false.
* `query_json` - (Required) A JSON describng the query according to the [Query Specification](https://docs.honeycomb.io/api/query-specification/#fields-on-a-query-specification). While the JSON can be constructed manually, it is easiest to use the `honeycombio_query` data source.
* `threshold` - (Required) A configuration block (described below) describing the threshold of the trigger.
* `frequency` - (Optional) The interval (in seconds) in which to check the results of the query’s calculation against the threshold. Value must be divisible by 60 and between 60 and 86400 (between 1 minute and 1 day). Defaults to 900 (15 minutes).
* `recipient` - (Optional) Zero or more configuration blocks (described below) with the recipients to notify when the trigger fires.

Each trigger configuration must contain exactly one `threshold` block, which accepts the following arguments:

* `op` - (Required) The operator to apply, allowed threshold operators are `>`, `>=`, `<`, and `<=`.
* `value` - (Required) The value to be used with the operator.

Each trigger configuration may have zero or more `recipient` blocks, which each accept the following arguments:

* `type` - (Required) The type of recipient, allowed types are `email`, `marker`, `pagerduty` and `slack`.
* `target` - (Optional) Target of the trigger, this has another meaning depending on the type of recipient (see the table below).
* `id` - (Optional) The ID of the recipient, this is necessary when type is Slack (see the note below).

Type        | Target
------------|-------------------------
`email`     | an email address
`marker`    | name of the marker
`pagerduty` | _N/A_
`slack`     | name of the channel

~> **NOTE** When type is Slack you have to specify the ID. Refer to [Specifying Recipients](https://docs.honeycomb.io/api/triggers/#specifying-recipients) for more information. It's currently not possible to retrieve this ID using the Terraform provider.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the trigger.
