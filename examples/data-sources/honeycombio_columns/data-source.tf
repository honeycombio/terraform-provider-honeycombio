variable "dataset" {
  type = string
}

# returns all columns
data "honeycombio_columns" "all" {
  dataset = var.dataset
}

# only returns the columns starting with 'foo_'
data "honeycombio_columns" "foo" {
  dataset     = var.dataset
  starts_with = "foo_"
}
