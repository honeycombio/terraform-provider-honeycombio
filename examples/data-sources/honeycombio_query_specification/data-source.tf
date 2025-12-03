data "honeycombio_query_specification" "example" {
  # zero or more calculation blocks
  calculation {
    op     = "AVG"
    column = "duration_ms"
  }

  calculated_field {
    name       = "fast_enough"
    expression = "LTE($response.duration_ms, 200)"
  }

  filter {
    column = "trace.parent_id"
    op     = "does-not-exist"
  }

  filter {
    column = "app.tenant"
    op     = "="
    value  = "ThatSpecialTenant" 
  }

  filter {
    column = "fast_enough"
    op     = "="
    value  = false
  }

  filter_combination = "AND"

  breakdowns = ["app.tenant"]
    
  time_range = 28800 // in seconds, 8 hours
  
  compare_time_offset = 86400 // in seconds, compare with data from 1 day ago
}

output "json_query" {
    value = data.honeycombio_query_specification.example.json
}
