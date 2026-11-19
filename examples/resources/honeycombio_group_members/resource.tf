data "honeycombio_user" "alice" {
  email = "alice@example.com"
}

data "honeycombio_user" "bob" {
  email = "bob@example.com"
}

resource "honeycombio_group" "payments" {
  name = "payments-oncall"
}

resource "honeycombio_group_members" "payments" {
  group_id = honeycombio_group.payments.id
  user_ids = [
    data.honeycombio_user.alice.id,
    data.honeycombio_user.bob.id,
  ]
}
