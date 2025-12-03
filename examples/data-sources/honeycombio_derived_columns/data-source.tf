variable "dataset" {
  type = string
}

# returns all columns
data "honeycombio_derived_columns" "all" {
  dataset = var.dataset
}

# only returns the derived columns starting with 'foo_'
data "honeycombio_derived_columns" "foo" {
  dataset     = var.dataset
  starts_with = "foo_"
}
