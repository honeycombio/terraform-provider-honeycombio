# Resource: honeycombio_trigger

Creates a trigger. For more information about triggers, check out [Alert with Triggers](https://docs.honeycomb.io/working-with-your-data/triggers/).

## Example Usage

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
}

resource "honeycombio_query" "example" {
  dataset    = var.dataset
  query_json = data.honeycombio_query_specification.example.json
}

resource "honeycombio_trigger" "example" {
  name        = "Requests are slower than usuals"
  description = "Average duration of all requests for the last 10 minutes."

  query_json = honeycombio_query.example.id
  dataset    = var.dataset

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
    type   = "marker"
    target = "Trigger - requests are slow"
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
* `frequency` - (Optional) The interval (in seconds) in which to check the results of the query’s calculation against the threshold. Value must be divisible by 60 and between 60 and 86400 (between 1 minute and 1 day). Defaults to 900 (15 minutes).
* `recipient` - (Optional) Zero or more configuration blocks (described below) with the recipients to notify when the trigger fires.

-> **NOTE** The query used in a Trigger must follow a strict subset: a query must contain exactly one calcuation and may only contain `calculation`, `filter`, `filter_combination` and `breakdowns` fields. The query's duration cannot be more than four times the trigger's frequency.

Each trigger configuration must contain exactly one `threshold` block, which accepts the following arguments:

* `op` - (Required) The operator to apply, allowed threshold operators are `>`, `>=`, `<`, and `<=`.
* `value` - (Required) The value to be used with the operator.

Each trigger configuration may have zero or more `recipient` blocks, which each accept the following arguments. A trigger recipient block can either refer to an existing recipient (a recipient that is already present in another trigger) or a new recipient. When specifying an existing recipient, only `id` may be set. If you pass in a recipient without its ID and only include the type and target, Honeycomb will make a best effort to match to an existing recipient. To retrieve the ID of an existing recipient, refer to the [`honeycombio_trigger_recipient`](../data-sources/trigger_recipient.md) data source.

* `type` - (Optional) The type of the trigger recipient, allowed types are `email`, `marker`, `pagerduty`, `slack` and `webhook`. Should not be used in combination with `id`.
* `target` - (Optional) Target of the trigger recipient, this has another meaning depending on the type of recipient (see the table below). Should not be used in combination with `id`.
* `id` - (Optional) The ID of an already existing recipient. Should not be used in combination with `type` and `target`.

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
