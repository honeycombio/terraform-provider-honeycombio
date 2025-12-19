variable "dataset" {
  type = string
}

# Create a flexible board first
resource "honeycombio_flexible_board" "example" {
  name        = "Service Monitoring Board"
  description = "A board for monitoring service health"
}

# Create a board view with various filter types
resource "honeycombio_board_view" "production_errors" {
  board_id = honeycombio_flexible_board.example.id
  name     = "Production Errors"

  filter {
    column    = "service.name"
    operation = "exists"
  }

  filter {
    column    = "environment"
    operation = "="
    value     = "production"
  }

  filter {
    column    = "status_code"
    operation = ">="
    value     = "500"
  }

  filter {
    column    = "error_type"
    operation = "in"
    value     = "timeout,database_error,network_error"
  }
}

# Another board view example with different filters
resource "honeycombio_board_view" "high_latency" {
  board_id = honeycombio_flexible_board.example.id
  name     = "High Latency Requests"

  filter {
    column    = "trace.parent_id"
    operation = "does-not-exist"
  }

  filter {
    column    = "duration_ms"
    operation = ">"
    value     = "1000"
  }

  filter {
    column    = "service.name"
    operation = "not-in"
    value     = "health-check,metrics"
  }
}

