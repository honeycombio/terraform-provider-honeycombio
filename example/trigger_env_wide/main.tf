terraform {
  required_providers {
    honeycombio = {
      source = "honeycombio/honeycombio"
    }
  }
}

provider "honeycombio" {
 api_key="GiBk40pefeEt8fk9qITAP2"
 api_url="http://localhost:8085"
}

data "honeycombio_query_specification" "query" {
  calculation {
    op     = "AVG"
    column = "db_dur_ms"
  }

  time_range = 900 // in seconds, 15 minutes
  
}


resource "honeycombio_trigger" "trigger" {
  name        = "Requests are slower than usual"
  description = "Average duration of all requests for ThatSpecialTenant for the last 15 minutes."

  disabled = false

  query_json = data.honeycombio_query_specification.query.json

  frequency = 900 // in seconds, 100 minutes

  threshold {
    op    = ">="
    value = 1000
  }

  # zero or more recipients
  recipient {
    type   = "email"
    target = "hello@example.com"
  }
  recipient {
    type   = "marker"
    target = "Trigger - slow requests" # name of the marker
  }

}


output "result" {
  value = resource.honeycombio_trigger.trigger
}