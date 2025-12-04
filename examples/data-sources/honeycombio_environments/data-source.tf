# returns all Environments
data "honeycombio_environments" "all" {}

# only returns the Environments starting with 'foo_'
data "honeycombio_environments" "foo" {
  detail_filter {
    name        = "name"
    value_regex = "foo_*"
  }
}
