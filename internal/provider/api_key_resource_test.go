package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAcc_APIKeyResource(t *testing.T) {
	t.Parallel()

	envID := testAccEnvironment()

	t.Run("happy path", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:                 testAccPreCheckV2API(t),
			ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
			Steps: []resource.TestStep{
				{
					Config: testAccConfigBasicAPIKeyTest("test key", "false", envID),
					Check: resource.ComposeAggregateTestCheckFunc(
						testAccEnsureAPIKeyExists(t, "honeycombio_api_key.test"),
						resource.TestCheckResourceAttrSet("honeycombio_api_key.test", "id"),
						resource.TestCheckResourceAttrSet("honeycombio_api_key.test", "secret"),
						resource.TestCheckResourceAttr("honeycombio_api_key.test", "name", "test key"),
						resource.TestCheckResourceAttr("honeycombio_api_key.test", "type", "ingest"),
						resource.TestCheckResourceAttr("honeycombio_api_key.test", "environment_id", envID),
						resource.TestCheckResourceAttr("honeycombio_api_key.test", "disabled", "false"),
						resource.TestCheckResourceAttr("honeycombio_api_key.test", "permissions.0.create_datasets", "true"),
					),
				},
				{ // now update the name and disabled state
					Config: testAccConfigBasicAPIKeyTest("updated test key", "true", envID),
					Check: resource.ComposeAggregateTestCheckFunc(
						testAccEnsureAPIKeyExists(t, "honeycombio_api_key.test"),
						resource.TestCheckResourceAttrSet("honeycombio_api_key.test", "id"),
						resource.TestCheckResourceAttrSet("honeycombio_api_key.test", "secret"),
						resource.TestCheckResourceAttr("honeycombio_api_key.test", "name", "updated test key"),
						resource.TestCheckResourceAttr("honeycombio_api_key.test", "type", "ingest"),
						resource.TestCheckResourceAttr("honeycombio_api_key.test", "environment_id", envID),
						resource.TestCheckResourceAttr("honeycombio_api_key.test", "disabled", "true"),
						resource.TestCheckResourceAttr("honeycombio_api_key.test", "permissions.0.create_datasets", "true"),
					),
				},
			},
		})
	})

	t.Run("default values", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:                 testAccPreCheckV2API(t),
			ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
			Steps: []resource.TestStep{
				{
					Config: fmt.Sprintf(`
resource "honeycombio_api_key" "test" {
  name = "test key"
  type = "ingest"

  environment_id = "%s"
}`, envID),
					Check: resource.ComposeAggregateTestCheckFunc(
						testAccEnsureAPIKeyExists(t, "honeycombio_api_key.test"),
						resource.TestCheckResourceAttrSet("honeycombio_api_key.test", "id"),
						resource.TestCheckResourceAttrSet("honeycombio_api_key.test", "secret"),
						resource.TestCheckResourceAttr("honeycombio_api_key.test", "name", "test key"),
						resource.TestCheckResourceAttr("honeycombio_api_key.test", "type", "ingest"),
						resource.TestCheckResourceAttr("honeycombio_api_key.test", "environment_id", envID),
						resource.TestCheckResourceAttr("honeycombio_api_key.test", "disabled", "false"),
					),
				},
			},
		})
	})
}

func testAccConfigBasicAPIKeyTest(name, disabled, envID string) string {
	return fmt.Sprintf(`
resource "honeycombio_api_key" "test" {
  name     = "%s"
  type     = "ingest"
  disabled = %s

  environment_id = "%s"

  permissions {
    create_datasets = true
  }
}`, name, disabled, envID)
}

func testAccEnsureAPIKeyExists(t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("resource not found: %s", name)
		}

		client := testAccV2Client(t)
		_, err := client.APIKeys.Get(context.Background(), rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("failed to fetch created API key: %s", err)
		}

		return nil
	}
}
