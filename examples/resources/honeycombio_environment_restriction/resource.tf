resource "honeycombio_group" "sre" {
  name = "sre"
}

data "honeycombio_environment" "production" {
  detail_filter {
    name  = "name"
    value = "production"
  }
}

resource "honeycombio_group_grant" "production" {
  group_id       = honeycombio_group.sre.id
  environment_id = data.honeycombio_environment.production.id
}

# grant first, then restrict, so the Group never loses access
resource "honeycombio_environment_restriction" "production" {
  environment_id = data.honeycombio_environment.production.id

  depends_on = [honeycombio_group_grant.production]
}
