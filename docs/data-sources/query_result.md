# Data Source: honeycombio_query_result

The `query_result` data source allows you to execute Honeycomb queries via the [Query Data API](https://docs.honeycomb.io/api/query-results/).
As this data source is a wrapper around the Query Data API all of its [documented restrictions](https://docs.honeycomb.io/api/query-results/#api-restrictions) apply.

~> **NOTE:** Use of this data source requires a Honeycomb Enterprise plan.

## Example Usage

```hcl
data "honeycombio_query_specification" "example" {
  time_range = 7200

  calculation {
    op = "COUNT"
  }
}

resource "honeycombio_query" "example" {
  dataset    = var.dataset
  query_json = data.honeycombio_query_specification.example.json
}

data "honeycombio_query_result" "example" {
  dataset  = var.dataset
  query_id = honeycombio_query.example.id
}

output "event_count" {
  value = format(
    "There have been %d events in the last %d seconds.",
    data.honeycombio_query_result.example.results[0]["COUNT"],
    data.honeycombio_query_specification.example.time_range
  )
}
```

~> **NOTE:** This data source is experimental and we're actively looking to learn how you are using it! Please consider opening an issue with feedback or joining the conversation in `#terraform-provider` in the [Pollinators Slack Community](https://join.slack.com/t/honeycombpollinators/shared_invite/zt-xqexg936-dckd0l29wdE3WLmUs8Qvpg).

## Argument Reference

The following arguments are supported:

* `dataset` - (Required) The dataset this query is associated with.
* `query_id` - (Required) The ID of the query that will be executed to obtain the result.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `query_url` - The permalink to the executed query's results.
* `query_image_url` - The permalink to the visualization of the executed query's results.
* `results` - The results of the executed query. This will be a list of maps, with each map's keys set to the breakdowns and calculations of the query. Due to a limitation of the Terraform Plugin SDK, all values are transformed into strings.
