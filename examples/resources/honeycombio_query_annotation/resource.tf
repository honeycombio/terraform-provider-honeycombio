variable "dataset" {
  type = string
}

data "honeycombio_query_specification" "test_query" {
  calculation {
    op     = "AVG"
    column = "duration_ms"
  }

  filter {
    column = "duration_ms"
    op     = ">"
    value  = 10
  }
}

resource "honeycombio_query" "test_query" {
  dataset    = var.dataset
  query_json = data.honeycombio_query_specification.test_query.json
}

resource "honeycombio_query_annotation" "test_annotation" {
	dataset     = var.dataset
	query_id    = honeycombio_query.test_query.id
	name        = "My Cool Query"
	description = "Describes my cool query (optional)"
}
