package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/test"
)

func TestAcc_GroupResource(t *testing.T) {
	name := test.RandomStringWithPrefix("test.", 20)
	newName := test.RandomStringWithPrefix("test.", 20)

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheckV2API(t),
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactory,
		CheckDestroy:             testAccEnsureGroupDestroyed(t),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigGroup(name, `description = "Reads payments"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccEnsureGroupExists(t, "honeycombio_group.test"),
					resource.TestCheckResourceAttrSet("honeycombio_group.test", "id"),
					resource.TestCheckResourceAttr("honeycombio_group.test", "name", name),
					resource.TestCheckResourceAttr("honeycombio_group.test", "description", "Reads payments"),
				),
			},
			{ // a rename is an update, not a replacement
				Config: testAccConfigGroup(newName, `description = "Reads payments and billing"`),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("honeycombio_group.test", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccEnsureGroupExists(t, "honeycombio_group.test"),
					resource.TestCheckResourceAttr("honeycombio_group.test", "name", newName),
					resource.TestCheckResourceAttr("honeycombio_group.test", "description", "Reads payments and billing"),
				),
			},
			{
				ResourceName:      "honeycombio_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{ // removing the description clears it
				Config: testAccConfigGroup(newName, ""),
				Check:  resource.TestCheckResourceAttr("honeycombio_group.test", "description", ""),
			},
			{ // a Group deleted outside Terraform is created again
				PreConfig: func() {
					c := testAccV2Client(t)
					pager, err := c.Groups.List(context.Background())
					if err != nil {
						t.Fatal(err)
					}
					for pager.HasNext() {
						groups, err := pager.Next(context.Background())
						if err != nil {
							t.Fatal(err)
						}
						for _, g := range groups {
							if g.Name == newName {
								if err := c.Groups.Delete(context.Background(), g.ID); err != nil {
									t.Fatal(err)
								}
							}
						}
					}
				},
				Config: testAccConfigGroup(newName, ""),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("honeycombio_group.test", plancheck.ResourceActionCreate),
					},
				},
				Check: testAccEnsureGroupExists(t, "honeycombio_group.test"),
			},
		},
	})
}

func testAccConfigGroup(name, extra string) string {
	return fmt.Sprintf(`
resource "honeycombio_group" "test" {
  name = "%s"
  %s
}`, name, extra)
}

func testAccEnsureGroupExists(t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("resource not found: %s", name)
		}

		g, err := testAccV2Client(t).Groups.Get(context.Background(), rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("failed to fetch created Group: %w", err)
		}
		if g.Name != rs.Primary.Attributes["name"] {
			return fmt.Errorf("Group name is %q in the API but %q in state", g.Name, rs.Primary.Attributes["name"])
		}
		return nil
	}
}

func testAccEnsureGroupDestroyed(t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "honeycombio_group" {
				continue
			}
			if _, err := testAccV2Client(t).Groups.Get(context.Background(), rs.Primary.ID); err == nil {
				return fmt.Errorf("Group %s still exists", rs.Primary.ID)
			}
		}
		return nil
	}
}
