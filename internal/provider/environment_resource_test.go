package provider

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAcc_EnvironmentResource(t *testing.T) {

	t.Run("happy path", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:                 testAccPreCheckV2API(t),
			ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
			Steps: []resource.TestStep{
				{
					Config: testAccConfigBasicEnvironmentTest("test environment", "test description", "blue", true),
					Check: resource.ComposeAggregateTestCheckFunc(
						testAccEnsureEnvironmentExists(t, "honeycombio_environment.test"),
						resource.TestCheckResourceAttrSet("honeycombio_environment.test", "id"),
						resource.TestCheckResourceAttrSet("honeycombio_environment.test", "slug"),
						resource.TestCheckResourceAttr("honeycombio_environment.test", "name", "test environment"),
						resource.TestCheckResourceAttr("honeycombio_environment.test", "description", "test description"),
						resource.TestCheckResourceAttr("honeycombio_environment.test", "color", "blue"),
						resource.TestCheckResourceAttr("honeycombio_environment.test", "delete_protected", "true"),
					),
				},
				{ // fail to delete protected environment
					Config:      testAccConfigBasicEnvironmentTest("test environment", "test description", "blue", true),
					Destroy:     true,
					ExpectError: regexp.MustCompile(`must disable delete protection`),
				},
				{ // now update the description and color, and disable delete protection
					Config: testAccConfigBasicEnvironmentTest("test environment", "new description", "green", false),
					Check: resource.ComposeAggregateTestCheckFunc(
						testAccEnsureEnvironmentExists(t, "honeycombio_environment.test"),
						resource.TestCheckResourceAttrSet("honeycombio_environment.test", "id"),
						resource.TestCheckResourceAttrSet("honeycombio_environment.test", "slug"),
						resource.TestCheckResourceAttr("honeycombio_environment.test", "name", "test environment"),
						resource.TestCheckResourceAttr("honeycombio_environment.test", "description", "new description"),
						resource.TestCheckResourceAttr("honeycombio_environment.test", "color", "green"),
						resource.TestCheckResourceAttr("honeycombio_environment.test", "delete_protected", "false"),
					),
				},
				{
					ResourceName: "honeycombio_environment.test",
					ImportState:  true,
				},
			},
		})
	})

	t.Run("default values", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:                 testAccPreCheckV2API(t),
			ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
			Steps: []resource.TestStep{
				{
					Config: `
resource "honeycombio_environment" "test" {
  name = "default"
}`,
					Check: resource.ComposeAggregateTestCheckFunc(
						testAccEnsureEnvironmentExists(t, "honeycombio_environment.test"),
						resource.TestCheckResourceAttrSet("honeycombio_environment.test", "id"),
						resource.TestCheckResourceAttrSet("honeycombio_environment.test", "slug"),
						resource.TestCheckResourceAttrSet("honeycombio_environment.test", "color"),
						resource.TestCheckResourceAttr("honeycombio_environment.test", "name", "default"),
						resource.TestCheckResourceAttr("honeycombio_environment.test", "description", ""),
						resource.TestCheckResourceAttr("honeycombio_environment.test", "delete_protected", "true"),
					),
				},
				{ // disable delete protection to allow cleanup
					Config: `
resource "honeycombio_environment" "test" {
  name             = "default"
  delete_protected = false
}`,
				},
			},
		})
	})

	t.Run("fails to create environment with delete protection disabled", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:                 testAccPreCheckV2API(t),
			ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
			Steps: []resource.TestStep{
				{
					Config:      testAccConfigBasicEnvironmentTest("nope", "", "red", false),
					PlanOnly:    true,
					ExpectError: regexp.MustCompile(`protection cannot be disabled at creation`),
				},
			},
		})
	})
}

func testAccConfigBasicEnvironmentTest(name, description, color string, protected bool) string {
	return fmt.Sprintf(`
resource "honeycombio_environment" "test" {
  name        = "%s"
  description = "%s"
  color       = "%s"

  delete_protected = %t
}`, name, description, color, protected)
}

func testAccEnsureEnvironmentExists(t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("resource not found: %s", name)
		}

		client := testAccV2Client(t)
		_, err := client.Environments.Get(context.Background(), rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("failed to fetch created Environment: %s", err)
		}

		return nil
	}
}
