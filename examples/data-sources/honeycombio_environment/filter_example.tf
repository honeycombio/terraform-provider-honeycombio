data "honeycombio_environment" "classic" {
  detail_filter {
    name  = "name"
    value = "Classic"
  }
}

data "honeycombio_environment" "prod" {
  detail_filter {
    name  = "name"
    value = "prod"
  }
}
