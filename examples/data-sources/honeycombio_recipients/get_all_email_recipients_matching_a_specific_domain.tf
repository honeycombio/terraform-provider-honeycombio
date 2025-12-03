data "honeycombio_recipients" "example-dot-com" {
  type = "email"

  detail_filter {
    name        = "address"
    value_regex = ".*@example.com"
  }
}
