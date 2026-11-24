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

# the Group has no Dataset grants, because dataset_ids is not set
resource "honeycombio_group_grants" "sre" {
  group_id        = honeycombio_group.sre.id
  environment_ids = [for e in data.honeycombio_environment.prod : e.id]
}
