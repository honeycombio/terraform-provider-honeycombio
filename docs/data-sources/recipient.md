# Data Source: honeycombio_recipient

Search the triggers or burn alerts of a dataset for a recipient. The ID of the existing recipient can be used when adding recipients to new triggers or burn alerts.

## Example Usage

```hcl
variable "dataset" {
  type = string
}

# search for a recipient of type "slack" and target "honeycomb-triggers" in the given dataset
data "honeycombio_recipient" "slack" {
  dataset = var.dataset
  type    = "slack"
  target  = "honeycomb-triggers"
}

data "honeycombio_query_specification" "example" {
  calculation {
    op     = "AVG"
    column = "duration_ms"
  }
}

resource "honeycombio_trigger" "example" {
  name        = "Requests are slower than usual"

  query_json = data.honeycombio_query_specification.example.json
  dataset    = var.dataset

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
    id = data.honeycombio_recipient.slack.id
  }
}
```

## Argument Reference

The following arguments are supported:

* `dataset` - (Required) The dataset this recipient is associated with.
* `type` - (Required) The type of recipient, allowed types are `email`, `marker`, `pagerduty`, `slack` and `webhook`.
* `target` - (Optional) Target of the trigger or burn alert, this has another meaning depending on the type of recipient (see the table below).

Type      | Target
----------|-------------------------
email     | an email address
marker    | name of the marker
pagerduty | _N/A_
slack     | name of the channel
webhook   | name of the webhook

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the recipient.
