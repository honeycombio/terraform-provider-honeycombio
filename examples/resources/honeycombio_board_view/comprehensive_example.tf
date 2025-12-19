variable "dataset" {
  type = string
}

# Create a flexible board
resource "honeycombio_flexible_board" "monitoring" {
  name        = "Application Monitoring"
  description = "Comprehensive monitoring dashboard"

  tags = {
    team    = "platform"
    project = "monitoring"
  }
}

# Board view for API errors
resource "honeycombio_board_view" "api_errors" {
  board_id = honeycombio_flexible_board.monitoring.id
  name     = "API Errors"

  filter {
    column    = "service.name"
    operation = "exists"
  }

  filter {
    column    = "http.status_code"
    operation = ">="
    value     = "400"
  }

  filter {
    column    = "environment"
    operation = "="
    value     = "production"
  }

  filter {
    column    = "error.message"
    operation = "contains"
    value     = "timeout"
  }
}

# Board view for specific services
resource "honeycombio_board_view" "core_services" {
  board_id = honeycombio_flexible_board.monitoring.id
  name     = "Core Services"

  filter {
    column    = "service.name"
    operation = "in"
    value     = "api-service,payment-service,user-service"
  }

  filter {
    column    = "duration_ms"
    operation = "<"
    value     = "500"
  }
}

# Board view for slow queries
resource "honeycombio_board_view" "slow_queries" {
  board_id = honeycombio_flexible_board.monitoring.id
  name     = "Slow Database Queries"

  filter {
    column    = "query.duration_ms"
    operation = ">"
    value     = "1000"
  }

  filter {
    column    = "database.name"
    operation = "exists"
  }

  filter {
    column    = "query.type"
    operation = "!="
    value     = "SELECT"
  }
}

