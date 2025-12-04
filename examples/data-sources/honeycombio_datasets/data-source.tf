# returns all datasets
data "honeycombio_datasets" "all" {}

# only returns the datasets with names starting with 'foo_'
data "honeycombio_datasets" "foo" {
  detail_filter {
    name        = "name"
    value_regex = "foo_*"
  }
}
