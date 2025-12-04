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

    frequency = 600 // in seconds, 10 minutes

    threshold {
        op             = ">="
        value          = 1000
    }
}
