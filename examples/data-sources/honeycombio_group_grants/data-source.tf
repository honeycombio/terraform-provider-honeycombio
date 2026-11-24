resource "honeycombio_group" "payments" {
  name = "payments-oncall"
}

data "honeycombio_group_grants" "payments" {
  group_id = honeycombio_group.payments.id
}

output "payments_dataset_ids" {
  value = data.honeycombio_group_grants.payments.dataset_ids
}

output "payments_environment_ids" {
  value = data.honeycombio_group_grants.payments.environment_ids
}
