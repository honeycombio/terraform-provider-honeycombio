# Resource: honeycombio_environment

Creates a Honeycomb Environment.

-> **NOTE** This resource requires the provider be configured with a Management Key with `environments:write` in the configured scopes.

## Example Usage

```hcl
resource "honeycombio_environment" "uat" {
  name  = "UAT-1"
  color = "green"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the Environment. Must be unique to the Team.
* `description` - (Optional) A description for the Environment.
* `color` - (Optional) The color to display the Environment in the navigation bar.
  If not provided one will be randomly selected at creation.
  One of `blue`, `green`, `gold`, `red`, `purple`, `lightBlue`, `lightGreen`, `lightGold`, `lightRed`, `lightPurple`.
* `delete_protected` - (Optional) the current state of the Environment's deletion protection status.
  Defaults to `true`. Cannot be set to `false` on create.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the Environment.
* `slug` - The slug of the Environment.

## Import

Environments can be imported by their ID. e.g.

```
$ terraform import honeycombio_environment.myenv hcaen_01j1jrsewaha3m0z6fwffpcrxg
```
