data "honeycombio_query_specification" "example" {
  time_range = 7200

  calculation {
    op = "COUNT"
  }
}

data "honeycombio_query_result" "example" {
  dataset    = var.dataset
  query_json = data.honeycombio_query_specification.example.json
}

output "event_count" {
  value = format(
    "There have been %d events in the last %d seconds.",
    data.honeycombio_query_result.example.results[0]["COUNT"],
    data.honeycombio_query_specification.example.time_range
  )
}
