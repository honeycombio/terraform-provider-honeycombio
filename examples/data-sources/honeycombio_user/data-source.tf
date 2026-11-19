variable "payments_emails" {
  type    = set(string)
  default = ["alice@example.com", "bob@example.com"]
}

# one lookup per email: the plan fails if any email matches no user on the Team
data "honeycombio_user" "payments" {
  for_each = var.payments_emails
  email    = each.value
}

resource "honeycombio_group" "payments" {
  name = "payments-oncall"
}

resource "honeycombio_group_members" "payments" {
  group_id = honeycombio_group.payments.id
  user_ids = [for u in data.honeycombio_user.payments : u.id]
}
