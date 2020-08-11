# Data Source: honeycombio_trigger_recipient

Search the triggers of a dataset for a trigger recipient. The ID of the already existing trigger recipient can be used when creating new triggers. Specifying a trigger recipient by ID is necessary when creating Slack recipients using the API.

## Example Usage

```hcl
variable "dataset" {
  type = string
}

# search for a trigger recipient of type "slack" and target "honeycomb-triggers" in the given dataaset
data "honeycombio_trigger_recipient" "slack" {
  dataset = var.dataset
  type    = "slack"
  target  = "honeycomb-triggers"
}

data "honeycombio_query" "example" {
  calculation {
    op     = "AVG"
    column = "duration_ms"
  }
}

resource "honeycombio_trigger" "example" {
  name        = "Requests are slower than usuals"

  query_json = data.honeycombio_query.example.json
  dataset    = var.dataset

  frequency = 600 // in seconds, 10 minutes

  threshold {
    op    = ">"
    value = 1000
  }

  recipient {
    type   = "email"
    target = "hello@example.com"
  }

  # add an already existing recipient
  recipient {
    id = data.honeycombio_trigger_recipient.slack
  }
}
```

## Argument Reference

The following arguments are supported:

* `dataset` - (Required) Search through all triggers linked to this dataset.
* `type` - (Required) The type of recipient, allowed types are `email`, `marker`, `pagerduty` and `slack`.
* `target` - (Optional) Target of the trigger, this has another meaning depending on the type of recipient (see the table below).

Type      | Target
----------|-------------------------
email     | an email address
marker    | name of the marker
pagerduty | _N/A_
slack     | name of the channel

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the trigger recipient.
