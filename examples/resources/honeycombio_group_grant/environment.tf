resource "honeycombio_group" "sre" {
  name = "sre"
}

data "honeycombio_environment" "production" {
  detail_filter {
    name  = "name"
    value = "production"
  }
}

# every Dataset in the Environment, including Datasets added later
resource "honeycombio_group_grant" "production" {
  group_id       = honeycombio_group.sre.id
  environment_id = data.honeycombio_environment.production.id
}
