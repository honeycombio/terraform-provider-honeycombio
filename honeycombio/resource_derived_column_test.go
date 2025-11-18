package honeycombio

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/test"
)

func TestAccHoneycombioDerivedColumn_basic(t *testing.T) {
	dataset := testAccDataset()
	alias := test.RandomStringWithPrefix("test.", 10)

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "honeycombio_derived_column" "test" {
  alias       = "%s"
  expression  = "BOOL(1)"
  description = "my test description"

  dataset = "%s"
}`, alias, dataset),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("honeycombio_derived_column.test", "alias", alias),
					resource.TestCheckResourceAttr("honeycombio_derived_column.test", "expression", "BOOL(1)"),
					resource.TestCheckResourceAttr("honeycombio_derived_column.test", "description", "my test description"),
				),
			},
			{
				ResourceName:      "honeycombio_derived_column.test",
				ImportStateId:     fmt.Sprintf("%s/%s", dataset, alias),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "honeycombio_derived_column" "invalid_column_in_expression" {
  alias       = "%s"
  expression  = "LOG10($invalid_column)"

  dataset = "%s"
}`, test.RandomStringWithPrefix("test.", 10), dataset),
				ExpectError: regexp.MustCompile(`unknown column`),
			},
		},
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
		Steps: []resource.TestStep{
			{
				Config: `
resource "honeycombio_derived_column" "test" {
  alias      = "invalid_syntax"
  expression = "BOOL(1"

  dataset = "foobar"
}`,
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`invalid derived column syntax: mismatched input '<EOF>'`),
			},
			{
				Config: `
resource "honeycombio_derived_column" "test" {
  alias      = "invalid_syntax"
  expression = "FOOBAR(1)"

  dataset = "foobar"
}`,
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`invalid derived column syntax: invalid function: FOOBAR`),
			},
			{
				Config: `
resource "honeycombio_derived_column" "test" {
  alias      = "invalid_syntax"
  expression = <<EOF
IF(AND(NOT(EXISTS($trace.parent_id)),EXISTS($duration_ms)),LTE($duration_ms,300)),
EOF

  dataset = "foobar"
}`,
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`invalid derived column syntax: extraneous input ','`),
			},
			{
				Config: `
resource "honeycombio_derived_column" "test" {
  alias      = "valid_syntax"
  expression = <<EOF
IF(AND(NOT(EXISTS($trace.parent_id)),EXISTS($duration_ms)),LTE($duration_ms,300))
EOF

  dataset = "foobar"
}`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				Config: `
resource "honeycombio_derived_column" "test_infix" {
  alias      = "valid_infix_syntax"
  expression = <<EOF
IF(!EXISTS($trace.parent_id) AND EXISTS($duration_ms), $duration_ms <= 300)
EOF

  dataset = "foobar"
}`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccHoneycombioDerivedColumn_AllToUnset(t *testing.T) {
	ctx := context.Background()
	c := testAccClient(t)

	if c.IsClassic(ctx) {
		t.Skip("env-wide Derived Columns are not supported in classic")
	}

	alias := test.RandomStringWithPrefix("test.", 10)

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "honeycombio_derived_column" "test" {
  alias       = "%s"
  expression  = "BOOL(1)"
  dataset     = "__all__"
}`, alias),
			},
			{
				Config: fmt.Sprintf(`
resource "honeycombio_derived_column" "test" {
  alias       = "%s"
  expression  = "BOOL(1)"
}`, alias),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}
