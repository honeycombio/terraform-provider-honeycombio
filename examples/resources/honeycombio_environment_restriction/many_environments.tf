resource "honeycombio_group" "sre" {
  name = "sre"
}

# the plan fails if a name matches no Environment
data "honeycombio_environment" "prod" {
  for_each = toset(["prod-us", "prod-eu"])

  detail_filter {
    name  = "name"
    value = each.value
  }
}

resource "honeycombio_group_grant" "prod" {
  for_each = data.honeycombio_environment.prod

  group_id       = honeycombio_group.sre.id
  environment_id = each.value.id
}

resource "honeycombio_environment_restriction" "prod" {
  for_each = data.honeycombio_environment.prod

  environment_id = each.value.id

  depends_on = [honeycombio_group_grant.prod]
}
