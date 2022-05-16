terraform {
  required_providers {
    honeycombio = {
      source = "honeycombio/honeycombio"
    }
  }
}

variable "dataset" {
  type = string
}

data "honeycombio_query_specification" "example" {
  time_range = 7 * 86400 # last 7 days

  calculation {
    op = "COUNT"
  }

  breakdowns = ["service.name"]
}


data "honeycombio_query_result" "example" {
  dataset    = var.dataset
  query_json = data.honeycombio_query_specification.example.json
}

locals {
  active_services = flatten(
    [
      for result in data.honeycombio_query_result.example.results : result["service.name"]
    ]
  )
}

# generate a board for each service seen in the last 7 days
resource "honeycombio_board" "services" {
  for_each = toset(local.active_services)

  name = format("%s board", each.key)
}
