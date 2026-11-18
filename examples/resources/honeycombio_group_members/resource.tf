resource "honeycombio_group" "payments" {
  name = "payments-oncall"
}

resource "honeycombio_group_members" "payments" {
  group_id = honeycombio_group.payments.id
  user_ids = [
    "hcxus_01jbr2v3k8m4n5p6q7r8s9t0vw",
    "hcxus_01jbr2w9z1a2b3c4d5e6f7g8hj",
  ]
}
