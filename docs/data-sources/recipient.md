# Data Source: honeycombio_recipient

`honeycombio_recipient` data source provides details about a specific recipient in the Team.

The ID of an existing recipient can be used when adding recipients to triggers or burn alerts.

-> **Note** Terraform will fail unless exactly one recipient is returned by the search. Ensure that your search is specific enough to return a single recipient ID only.
If you want to match multiple recipients, use the `honeycombio_recipients` data source instead.

## Example Usage

```hcl
# search for a Slack recipient with channel name "honeycomb-triggers"
data "honeycombio_recipient" "slack" {
  type    = "slack"

  detail_filter {
    name  = "channel"
    value = "#honeycomb-triggers"
  }
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

* `type` - (Required) The type of recipient, allowed types are `email`, `pagerduty`, `msteams`, `slack` and `webhook`.
* `dataset` - (Optional) Deprecated: recipients are now a Team-level construct. Any provided value will be ignored.
* `detail_filter` - (Optional) a block to further filter recipients as described below.
* `target` - (Optional) Deprecated: use `detail_filter` instead. The target of the recipient, this has another meaning depending on the type of recipient (see the table below).

Type      | Target
----------|-------------------------
email     | an email address
marker    | name of the marker
msteams   | name of the integration
pagerduty | _N/A_
slack     | name of the channel
webhook   | name of the webhook

To further filter the recipient results, a `detail_filter` block can be provided which accepts the following arguments:

* `name` - (Required) The name of the detail field to filter by. Allowed values are `address`, `channel`, `name`, `integration_name`, and `url`.
* `value` - (Optional) The value of the detail field to match on.
* `value_regex` - (Optional) A regular expression string to apply to the value of the detail field to match on.

~> **Note** one of `value` or `value_regex` is required.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the recipient.
* `address` - The email recipient's address -- if of type `email`.
* `channel` - The Slack recipient's channel -- if of type `slack`.
* `name` - The webhook recipient's name -- if of type `webhook` or `msteams`.
* `secret` - (Sensitive) The webhook recipient's secret -- if of type `webhook`.
* `url` - The webhook recipient's URL - if of type `webhook` or `msteams`.
* `integration_key` - (Sensitive) The PagerDuty recipient's integration key -- if of type `pagerduty`.
* `integration_name` - The PagerDuty recipient's inregration name -- if of type `pagerduty`.
