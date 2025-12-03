variable "dataset" {
    type = string
}

data "honeycombio_query_specification" "example" {
    calculation {
        op     = "AVG"
        column = "duration_ms"
    }
}

resource "honeycombio_trigger" "example" {
    name        = "Requests are slower than usual"
    description = "Average duration of all requests for the last 10 minutes."

    query_json = data.honeycombio_query_specification.example.json
    dataset    = var.dataset

    frequency = 600 // in seconds, 10 minutes

    threshold {
        op             = ">="
        value          = 1000
    }

    baseline_details {
        type            = "percentage"
        offset_minutes  = 1440
    }

    tags = {
        team = "backend"
        env  = "production"
    }
}
