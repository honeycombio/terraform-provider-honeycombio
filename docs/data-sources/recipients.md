# Data Source: honeycombio_recipients

`honeycombio_recipients` data source provides recipient IDs of recipients matching a set of criteria.

## Example Usage

### Get all recipients
```hcl
data "honeycombio_recipients" "all" {}
```

### Get all email recipients matching a specific domain
```hcl
data "honeycombio_recipients" "example-dot-com" {
  type = "email"

  detail_filter {
    name        = "address"
    value_regex = ".*@example.com"
  }
}
```

## Argument Reference

The following arguments are supported:

* `type` - (Optional) The type of recipient, allowed types are `email`, `pagerduty`, `slack` and `webhook`.
* `detail_filter` - (Optional) a block to further filter recipients as described below. `type` must be set when providing a filter.

To further filter the recipient results, a `filter_detail` block can be provided which accepts the following arguments:

* `name` - (Required) The name of the detail field to filter by. Allowed values are `address`, `channel`, `name`, `integration_name`, and `url`.
* `value` - (Optional) The value of the detail field to match on.
* `value_regex` - (Optional) A regular expression string to apply to the value of the detail field to match on.

~> **Note** one of `value` or `value_regex` is required.

## Attribute Reference

* `ids` - A list of all the recipient IDs found.
