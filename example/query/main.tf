terraform {
  required_providers {
    honeycombio = {
      source = "honeycombio/honeycombio"
    }
  }
}

variable "dataset" {
  type        = string
  description = "The dataset to create the query in"
}

# Create a query specification with compare_time_offset_seconds
data "honeycombio_query_specification" "comparison_query" {
  calculation {
    op = "COUNT"
  }

  # Filter for successful requests only
  filter {
    column = "status_code"
    op     = "="
    value  = "200"
  }

  # Group by service name
  breakdowns = ["service.name"]

  # Query the last 4 hours
  time_range = 14400 // 4 hours in seconds

  # Compare with data from 24 hours ago (1 day)
  compare_time_offset = 86400 // 24 hours in seconds

  # Limit results to top 10 services
  limit = 10

  # Order by average duration descending
  order {
    op     = "AVG"
    column = "duration_ms"
    order  = "descending"
  }
}

# Create the query using the specification
resource "honeycombio_query" "comparison_query" {
  dataset    = var.dataset
  query_json = data.honeycombio_query_specification.comparison_query.json
}

# Output the query ID and JSON for reference
output "query_id" {
  value       = honeycombio_query.comparison_query.id
  description = "The ID of the created query"
}

output "query_json" {
  value       = data.honeycombio_query_specification.comparison_query.json
  description = "The JSON specification of the query with comparison"
}
