variable "dataset" {
  type = string
}

# returns all SLOs
data "honeycombio_slos" "all" {
  dataset = var.dataset
}

# only returns the SLOs starting with 'foo_'
data "honeycombio_slos" "foo" {
  dataset     = var.dataset

  detail_filter {
    name        = "name"
    value_regex = "foo_*"
  }

  detail_filter {
    name     = "tags"
    operator = "contains"
    value    = "team:core"
  }
}
