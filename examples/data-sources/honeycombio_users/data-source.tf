variable "payments_emails" {
  type    = list(string)
  default = ["alice@example.com", "bob@example.com"]
}

# every user on the Team
data "honeycombio_users" "all" {}

locals {
  # The API returns emails in lower case, so compare in lower case.
  payments_emails = toset([for e in var.payments_emails : lower(e)])
}

resource "honeycombio_group" "payments" {
  name = "payments-oncall"
}

resource "honeycombio_group_members" "payments" {
  group_id = honeycombio_group.payments.id
  # emails that match no user on the Team are left out
  user_ids = [for u in data.honeycombio_users.all.users : u.id if contains(local.payments_emails, u.email)]
}

# only the users in one Group: read the Group through honeycombio_group_members,
# so the list is read after this apply changes the members
data "honeycombio_users" "payments" {
  group_id = honeycombio_group_members.payments.group_id
}

output "payments_emails" {
  value = data.honeycombio_users.payments.users[*].email
}
