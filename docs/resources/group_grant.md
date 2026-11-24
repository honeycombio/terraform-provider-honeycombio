# Resource: honeycombio_group_grant

Gives a Group access to one restricted Dataset, or to every Dataset in an Environment.
The members of the Group can read the Datasets that the grant names.

-> This resource requires the provider be configured with a Management Key with `groups:write` in the configured scopes.
  A grant with `environment_id` also needs `environments:read`.

The Group has access while this resource exists.

* Set exactly one of `dataset_id` and `environment_id`.
* Creating this resource fails for a grant that already exists. Import the grant instead.
* If the grant is revoked outside Terraform, the next apply grants it again.
* Do not use this resource on a Group that `honeycombio_group_grants` manages. That resource revokes every grant that it does not list.

## Dataset grants

The Dataset must be restricted. A grant on a Dataset that is not restricted fails.

Set `dataset_id` from the `honeycombio_dataset_restriction` resource, as the example shows.
Terraform then restricts the Dataset before it grants it, and revokes the grant before it removes the restriction.

~> **Note** When a Dataset loses its restriction, the API deletes every grant on it.
  If the restriction was removed outside Terraform, the next apply restricts the Dataset and grants it again.

## Environment grants

An Environment grant gives the Group access to every Dataset in the Environment, including Datasets added later.
The Environment does not have to be restricted. In an Environment that is not restricted, the grant gives access to its restricted Datasets.
The grant stays when the Environment is restricted or unrestricted. When the Environment is deleted, the API deletes the grant.

To restrict an Environment and keep the Group's access, grant the Group access first. Then restrict the Environment.

## Grants made outside Terraform

This resource manages only its own grant.
Terraform does not read the Group's other grants, such as grants made in the UI.
Choose what happens to them:

| To                         | Use                                                                                                                                               |
| -------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------- |
| Revoke them                | The `honeycombio_group_grants` resource instead of this one. It owns the Group's whole list of grants.                                            |
| Get a warning about them   | This resource and a `check` block with the `honeycombio_group_grants` data source, as the data source docs show. A warning never blocks an apply. |
| Keep them, with no warning | This resource only.                                                                                                                               |

## Example Usage

### One Dataset

```terraform
data "honeycombio_auth_metadata" "current" {}

resource "honeycombio_dataset" "checkout" {
  name = "checkout"
}

resource "honeycombio_dataset_restriction" "checkout" {
  environment_id = data.honeycombio_auth_metadata.current.environment.id
  dataset_id     = honeycombio_dataset.checkout.hcid
}

resource "honeycombio_group" "payments" {
  name = "payments-oncall"
}

# The reference to the restriction makes Terraform restrict the Dataset
# before it grants it, and revoke the grant before it removes the restriction.
resource "honeycombio_group_grant" "checkout" {
  group_id   = honeycombio_group.payments.id
  dataset_id = honeycombio_dataset_restriction.checkout.dataset_id
}
```

### One Environment

```terraform
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
```

### Several Environments

`for_each` makes one grant for each Environment.

```terraform
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
```

### Several Groups and several Datasets

`setproduct` makes one grant for each pair of a Group and a Dataset.
Each grant references its own restriction, so Terraform restricts each Dataset before it grants it.

```terraform
data "honeycombio_auth_metadata" "current" {}

locals {
  payments_datasets = toset(["checkout", "billing", "refunds"])
  payments_groups   = toset(["payments-oncall", "finance"])
}

data "honeycombio_dataset" "payments" {
  for_each = local.payments_datasets
  slug     = each.key
}

resource "honeycombio_dataset_restriction" "payments" {
  for_each       = data.honeycombio_dataset.payments
  environment_id = data.honeycombio_auth_metadata.current.environment.id
  dataset_id     = each.value.hcid
}

resource "honeycombio_group" "payments" {
  for_each = local.payments_groups
  name     = each.key
}

# One grant for each pair of a Group and a Dataset
resource "honeycombio_group_grant" "payments" {
  for_each = {
    for pair in setproduct(local.payments_groups, local.payments_datasets) :
    "${pair[0]}/${pair[1]}" => { group = pair[0], dataset = pair[1] }
  }
  group_id   = honeycombio_group.payments[each.value.group].id
  dataset_id = honeycombio_dataset_restriction.payments[each.value.dataset].dataset_id
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `group_id` (String) The ID of the Group to give access to.

### Optional

- `dataset_id` (String) The unique ID of the Dataset that the Group can read. It is the `hcid` of the `honeycombio_dataset` resource and data source, not the Dataset's slug. The Dataset must be restricted. Set exactly one of `dataset_id` and `environment_id`.
- `environment_id` (String) The ID of the Environment whose Datasets the Group can read, including Datasets added later. It is the `id` of the `honeycombio_environment` resource and data source, not the Environment's slug. The Environment does not have to be restricted. Set exactly one of `dataset_id` and `environment_id`.

### Read-Only

- `id` (String) The ID of the grant, written as `<group_id>/dataset/<dataset_id>` or `<group_id>/environment/<environment_id>`.

## Import

Group grants can be imported by the ID of the Group and the ID of the Dataset or the Environment.
Write a Dataset grant as `<group_id>/dataset/<dataset_id>`, and an Environment grant as `<group_id>/environment/<environment_id>`. e.g.

```
$ terraform import honeycombio_group_grant.checkout hcxgr_01jbr2v3k8m4n5p6q7r8s9t0vw/dataset/hcxds_01jbr4k2m7n8p9q0r1s2t3v4wx
$ terraform import honeycombio_group_grant.production hcxgr_01jbr2v3k8m4n5p6q7r8s9t0vw/environment/hcxen_01hznmeqrcq8rz533xrvtc6mk0
```

The import fails for a grant that does not exist.
