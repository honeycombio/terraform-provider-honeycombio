---
page_title: "Control who can read your data"
---

# Control who can read your data

This guide shows one configuration that controls who can read your data in Honeycomb.
The configuration creates a Group, adds users to it, restricts data, and gives the Group access to that data.
It uses both kinds of restriction: a restriction on single Datasets, and a restriction on a whole Environment.

-> This feature is available as part of the [Honeycomb Enterprise plan](https://www.honeycomb.io/pricing/).

## How access works

Only Team owners and the members of a Group that has access can read a restricted Dataset.
Every Team member can read a Dataset that is not restricted.

You can restrict data in two ways:

| To restrict                                                     | Use                                                                              | Grant the Group access with `honeycombio_group_grants` |
| --------------------------------------------------------------- | -------------------------------------------------------------------------------- | ------------------------------------------------------ |
| One Dataset                                                     | [`honeycombio_dataset_restriction`](../resources/dataset_restriction.md)         | `dataset_ids`                                          |
| Every Dataset in an Environment, including Datasets added later | [`honeycombio_environment_restriction`](../resources/environment_restriction.md) | `environment_ids`                                      |

A restricted Environment cannot have Dataset restrictions or Dataset grants.
For this reason, the example uses two Environments: a shared Environment with some restricted Datasets, and an Environment that holds only payments data.

## What the example builds

The example builds these objects, in this order:

1. A Group named `payments`, with two members.
2. A restriction on each of three Datasets in the shared Environment: `checkout`, `billing`, and `refunds`.
   Every Team member can still read the other Datasets in that Environment.
3. Access for the Group to the three Datasets and to the `payments-production` Environment.
4. A restriction on the `payments-production` Environment.

## Requirements

The configuration needs two keys:

* A Configuration Key for the shared Environment.
  The `honeycombio_dataset` and `honeycombio_auth_metadata` data sources use it.
* A Management Key with these scopes: `groups:write`, `users:read`, `environments:read`, and `environments:write`.

Set both keys in the provider, as the [provider documentation](../index.md) describes.

The configuration manages access only. These objects must already exist:

* The users, on the Team.
* The Datasets, in the shared Environment.
* The `payments-production` Environment. It must be a different Environment from the shared Environment.

## Example configuration

```terraform
terraform {
  required_providers {
    honeycombio = {
      source = "honeycombio/honeycombio"
    }
  }
}

locals {
  # The people who can read the payments data
  member_emails = toset(["alice@example.com", "bob@example.com"])

  # Payments Datasets in a shared Environment: the Environment of the
  # provider's Configuration Key. The other Datasets in it stay open.
  dataset_slugs = toset(["checkout", "billing", "refunds"])

  # An Environment that holds only payments data. Every Dataset in it is
  # restricted, including Datasets added later.
  environment_name = "payments-production"
}

# 1. The Group and its members

resource "honeycombio_group" "payments" {
  name        = "payments"
  description = "Engineers who can read the payments data."
}

# the plan fails if an email matches no user on the Team
data "honeycombio_user" "payments" {
  for_each = local.member_emails
  email    = each.key
}

resource "honeycombio_group_members" "payments" {
  group_id = honeycombio_group.payments.id
  user_ids = [for u in data.honeycombio_user.payments : u.id]
}

# 2. Restrict single Datasets in the shared Environment

data "honeycombio_auth_metadata" "current" {}

data "honeycombio_dataset" "payments" {
  for_each = local.dataset_slugs
  slug     = each.key
}

resource "honeycombio_dataset_restriction" "payments" {
  for_each       = data.honeycombio_dataset.payments
  environment_id = data.honeycombio_auth_metadata.current.environment.id
  dataset_id     = each.value.hcid
}

# 3. Give the Group access to the Datasets and to the payments Environment

data "honeycombio_environment" "payments" {
  detail_filter {
    name  = "name"
    value = local.environment_name
  }
}

# The references to the restrictions make Terraform restrict each Dataset
# before it grants it, and revoke the grant before it removes the restriction.
resource "honeycombio_group_grants" "payments" {
  group_id        = honeycombio_group.payments.id
  dataset_ids     = [for r in honeycombio_dataset_restriction.payments : r.dataset_id]
  environment_ids = [data.honeycombio_environment.payments.id]
}

# 4. Restrict the payments Environment

# depends_on makes Terraform grant before it restricts, so the Group never
# loses access. On destroy, Terraform unrestricts before it revokes.
resource "honeycombio_environment_restriction" "payments" {
  environment_id = data.honeycombio_environment.payments.id

  depends_on = [honeycombio_group_grants.payments]
}
```

## Order of the changes

Terraform orders the changes by the references between the resources:

* `dataset_ids` refers to the Dataset restrictions.
  Terraform restricts each Dataset before it grants the Dataset to the Group.
* The Environment restriction has the grants in `depends_on`.
  Terraform grants the Environment to the Group before it restricts the Environment, so the members of the Group never lose access.

On destroy, Terraform does the same steps in the opposite order.

## Keep data restricted

~> **Warning** When Terraform deletes a restriction, every Team member can read that data again.
  When a Dataset loses its restriction, the API also revokes every grant on the Dataset.

To stop an accidental delete, add `prevent_destroy` to the `lifecycle` block of each restriction.
Terraform then stops a plan that deletes the restriction.

## Manage grants

This example uses `honeycombio_group_grants`, which owns the Group's whole list of grants.
It revokes grants that the configuration does not list, such as grants made in the UI.

To keep grants made outside Terraform, use one [`honeycombio_group_grant`](../resources/group_grant.md) for each grant instead.
Use one of the two resources for each Group, never both.

## Restrict an Environment that has restricted Datasets

When an Environment becomes restricted, the API removes the restriction of each Dataset in it, and the grants on those Datasets.
To move from Dataset restrictions to an Environment restriction, do the steps in the order that [`honeycombio_environment_restriction`](../resources/environment_restriction.md) describes.
If you remove the Dataset restrictions first, every Team member can read those Datasets until the Environment is restricted.
