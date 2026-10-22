package provider

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"

	tfresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"
)

func TestAcc_APIKeyResource(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	c := testAccV2Client(t)
	env := testAccEnvironment(ctx, t, c)

	t.Run("happy path for ingest key", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:                 testAccPreCheckV2API(t),
			ProtoV6ProviderFactories: testAccProtoV6MuxServerFactory,
			Steps: []resource.TestStep{
				{
					Config: testAccConfigIngestAPIKeyTest("test key", "false", env.ID),
					Check: resource.ComposeAggregateTestCheckFunc(
						testAccEnsureAPIKeyExists(t, "honeycombio_api_key.test"),
						resource.TestCheckResourceAttrSet("honeycombio_api_key.test", "id"),
						resource.TestCheckResourceAttrSet("honeycombio_api_key.test", "secret"),
						resource.TestCheckResourceAttrSet("honeycombio_api_key.test", "key"),
						resource.TestCheckResourceAttr("honeycombio_api_key.test", "name", "test key"),
						resource.TestCheckResourceAttr("honeycombio_api_key.test", "type", "ingest"),
						resource.TestCheckResourceAttr("honeycombio_api_key.test", "environment_id", env.ID),
						resource.TestCheckResourceAttr("honeycombio_api_key.test", "disabled", "false"),
						resource.TestCheckResourceAttr("honeycombio_api_key.test", "permissions.0.create_datasets", "true"),
					),
				},
				{ // now update the name and disabled state
					Config: testAccConfigIngestAPIKeyTest("updated test key", "true", env.ID),
					Check: resource.ComposeAggregateTestCheckFunc(
						testAccEnsureAPIKeyExists(t, "honeycombio_api_key.test"),
						resource.TestCheckResourceAttrSet("honeycombio_api_key.test", "id"),
						resource.TestCheckResourceAttrSet("honeycombio_api_key.test", "secret"),
						resource.TestCheckResourceAttrSet("honeycombio_api_key.test", "key"),
						resource.TestCheckResourceAttr("honeycombio_api_key.test", "name", "updated test key"),
						resource.TestCheckResourceAttr("honeycombio_api_key.test", "type", "ingest"),
						resource.TestCheckResourceAttr("honeycombio_api_key.test", "environment_id", env.ID),
						resource.TestCheckResourceAttr("honeycombio_api_key.test", "disabled", "true"),
						resource.TestCheckResourceAttr("honeycombio_api_key.test", "permissions.0.create_datasets", "true"),
					),
				},
			},
		})
	})

	t.Run("happy path for configuration key", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:                 testAccPreCheckV2API(t),
			ProtoV6ProviderFactories: testAccProtoV6MuxServerFactory,
			Steps: []resource.TestStep{
				{
					Config: testAccConfigConfigurationAPIKeyTest("test key", "false", "false", env.ID),
					Check: resource.ComposeAggregateTestCheckFunc(
						testAccEnsureAPIKeyExists(t, "honeycombio_api_key.test"),
						resource.TestCheckResourceAttrSet("honeycombio_api_key.test", "id"),
						resource.TestCheckResourceAttrSet("honeycombio_api_key.test", "secret"),
						resource.TestCheckResourceAttrSet("honeycombio_api_key.test", "key"),
						resource.TestCheckResourceAttr("honeycombio_api_key.test", "name", "test key"),
						resource.TestCheckResourceAttr("honeycombio_api_key.test", "type", "configuration"),
						resource.TestCheckResourceAttr("honeycombio_api_key.test", "environment_id", env.ID),
						resource.TestCheckResourceAttr("honeycombio_api_key.test", "visible_to_members", "false"),
						resource.TestCheckResourceAttr("honeycombio_api_key.test", "disabled", "false"),
						// Check all the permissions to make sure they match
						resource.TestCheckResourceAttr("honeycombio_api_key.test", "permissions.0.send_events", "true"),
						resource.TestCheckResourceAttr("honeycombio_api_key.test", "permissions.0.create_datasets", "true"),
						resource.TestCheckResourceAttr("honeycombio_api_key.test", "permissions.0.manage_queries", "true"),
						resource.TestCheckResourceAttr("honeycombio_api_key.test", "permissions.0.run_queries", "true"),
						resource.TestCheckResourceAttr("honeycombio_api_key.test", "permissions.0.read_service_maps", "true"),
						resource.TestCheckResourceAttr("honeycombio_api_key.test", "permissions.0.manage_public_boards", "true"),
						resource.TestCheckResourceAttr("honeycombio_api_key.test", "permissions.0.manage_private_boards", "false"),
						resource.TestCheckResourceAttr("honeycombio_api_key.test", "permissions.0.manage_slos", "true"),
						resource.TestCheckResourceAttr("honeycombio_api_key.test", "permissions.0.manage_triggers", "true"),
						resource.TestCheckResourceAttr("honeycombio_api_key.test", "permissions.0.manage_recipients", "true"),
						resource.TestCheckResourceAttr("honeycombio_api_key.test", "permissions.0.manage_markers", "true"),
					),
				},
				{ // now update the name and disabled state
					Config: testAccConfigConfigurationAPIKeyTest("updated test key", "true", "false", env.ID),
					Check: resource.ComposeAggregateTestCheckFunc(
						testAccEnsureAPIKeyExists(t, "honeycombio_api_key.test"),
						resource.TestCheckResourceAttrSet("honeycombio_api_key.test", "id"),
						resource.TestCheckResourceAttrSet("honeycombio_api_key.test", "secret"),
						resource.TestCheckResourceAttrSet("honeycombio_api_key.test", "key"),
						resource.TestCheckResourceAttr("honeycombio_api_key.test", "name", "updated test key"),
						resource.TestCheckResourceAttr("honeycombio_api_key.test", "type", "configuration"),
						resource.TestCheckResourceAttr("honeycombio_api_key.test", "environment_id", env.ID),
						resource.TestCheckResourceAttr("honeycombio_api_key.test", "visible_to_members", "false"),
						resource.TestCheckResourceAttr("honeycombio_api_key.test", "disabled", "true"),
						// Check all the permissions to make sure they match
						resource.TestCheckResourceAttr("honeycombio_api_key.test", "permissions.0.send_events", "true"),
						resource.TestCheckResourceAttr("honeycombio_api_key.test", "permissions.0.create_datasets", "true"),
						resource.TestCheckResourceAttr("honeycombio_api_key.test", "permissions.0.manage_queries", "true"),
						resource.TestCheckResourceAttr("honeycombio_api_key.test", "permissions.0.run_queries", "true"),
						resource.TestCheckResourceAttr("honeycombio_api_key.test", "permissions.0.read_service_maps", "true"),
						resource.TestCheckResourceAttr("honeycombio_api_key.test", "permissions.0.manage_public_boards", "true"),
						resource.TestCheckResourceAttr("honeycombio_api_key.test", "permissions.0.manage_private_boards", "false"),
						resource.TestCheckResourceAttr("honeycombio_api_key.test", "permissions.0.manage_slos", "true"),
						resource.TestCheckResourceAttr("honeycombio_api_key.test", "permissions.0.manage_triggers", "true"),
						resource.TestCheckResourceAttr("honeycombio_api_key.test", "permissions.0.manage_recipients", "true"),
						resource.TestCheckResourceAttr("honeycombio_api_key.test", "permissions.0.manage_markers", "true"),
					),
				},
			},
		})
	})

	t.Run("default values", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:                 testAccPreCheckV2API(t),
			ProtoV6ProviderFactories: testAccProtoV6MuxServerFactory,
			Steps: []resource.TestStep{
				{
					Config: fmt.Sprintf(`
resource "honeycombio_api_key" "test_ingest" {
  name = "test key"
  type = "ingest"

  environment_id = "%s"
}`, env.ID),
					Check: resource.ComposeAggregateTestCheckFunc(
						testAccEnsureAPIKeyExists(t, "honeycombio_api_key.test_ingest"),
						resource.TestCheckResourceAttrSet("honeycombio_api_key.test_ingest", "id"),
						resource.TestCheckResourceAttrSet("honeycombio_api_key.test_ingest", "secret"),
						resource.TestCheckResourceAttr("honeycombio_api_key.test_ingest", "name", "test key"),
						resource.TestCheckResourceAttr("honeycombio_api_key.test_ingest", "description", ""),
						resource.TestCheckResourceAttr("honeycombio_api_key.test_ingest", "type", "ingest"),
						resource.TestCheckResourceAttr("honeycombio_api_key.test_ingest", "environment_id", env.ID),
						resource.TestCheckResourceAttr("honeycombio_api_key.test_ingest", "disabled", "false"),
					),
				},
				{
					Config: fmt.Sprintf(`
resource "honeycombio_api_key" "test_configuration" {
  name = "test key"
  type = "configuration"

  environment_id = "%s"

  permissions {}
}`, env.ID),
					Check: resource.ComposeAggregateTestCheckFunc(
						testAccEnsureAPIKeyExists(t, "honeycombio_api_key.test_configuration"),
						resource.TestCheckResourceAttrSet("honeycombio_api_key.test_configuration", "id"),
						resource.TestCheckResourceAttrSet("honeycombio_api_key.test_configuration", "secret"),
						resource.TestCheckResourceAttrSet("honeycombio_api_key.test_configuration", "key"),
						resource.TestCheckResourceAttr("honeycombio_api_key.test_configuration", "name", "test key"),
						resource.TestCheckResourceAttr("honeycombio_api_key.test_configuration", "description", ""),
						resource.TestCheckResourceAttr("honeycombio_api_key.test_configuration", "type", "configuration"),
						resource.TestCheckResourceAttr("honeycombio_api_key.test_configuration", "environment_id", env.ID),
						resource.TestCheckResourceAttr("honeycombio_api_key.test_configuration", "visible_to_members", "false"),
						resource.TestCheckResourceAttr("honeycombio_api_key.test_configuration", "disabled", "false"),
						// Check all the permissions to make sure they match
						resource.TestCheckResourceAttr("honeycombio_api_key.test_configuration", "permissions.0.send_events", "false"),
						resource.TestCheckResourceAttr("honeycombio_api_key.test_configuration", "permissions.0.create_datasets", "false"),
						resource.TestCheckResourceAttr("honeycombio_api_key.test_configuration", "permissions.0.manage_queries", "false"),
						resource.TestCheckResourceAttr("honeycombio_api_key.test_configuration", "permissions.0.run_queries", "false"),
						resource.TestCheckResourceAttr("honeycombio_api_key.test_configuration", "permissions.0.read_service_maps", "false"),
						resource.TestCheckResourceAttr("honeycombio_api_key.test_configuration", "permissions.0.manage_public_boards", "false"),
						resource.TestCheckResourceAttr("honeycombio_api_key.test_configuration", "permissions.0.manage_private_boards", "false"),
						resource.TestCheckResourceAttr("honeycombio_api_key.test_configuration", "permissions.0.manage_slos", "false"),
						resource.TestCheckResourceAttr("honeycombio_api_key.test_configuration", "permissions.0.manage_triggers", "false"),
						resource.TestCheckResourceAttr("honeycombio_api_key.test_configuration", "permissions.0.manage_recipients", "false"),
						resource.TestCheckResourceAttr("honeycombio_api_key.test_configuration", "permissions.0.manage_markers", "false"),
					),
				},
			},
		})
	})

	t.Run("description lifecycle", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:                 testAccPreCheckV2API(t),
			ProtoV6ProviderFactories: testAccProtoV6MuxServerFactory,
			Steps: []resource.TestStep{
				{
					Config: testAccConfigAPIKeyDescriptionTest(`description = "first description"`, env.ID),
					Check:  testAccCheckAPIKeyDescription(t, "first description"),
				},
				{ // change it, using characters outside of ASCII
					Config: testAccConfigAPIKeyDescriptionTest(`description = "changed \U0001F600"`, env.ID),
					Check:  testAccCheckAPIKeyDescription(t, "changed \U0001F600"),
				},
				{ // clear it with an empty string
					Config: testAccConfigAPIKeyDescriptionTest(`description = ""`, env.ID),
					Check:  testAccCheckAPIKeyDescription(t, ""),
				},
				{ // set it again
					Config: testAccConfigAPIKeyDescriptionTest(`description = "second description"`, env.ID),
					Check:  testAccCheckAPIKeyDescription(t, "second description"),
				},
				{ // clear it by removing the attribute
					Config: testAccConfigAPIKeyDescriptionTest(``, env.ID),
					Check:  testAccCheckAPIKeyDescription(t, ""),
				},
			},
		})
	})

	t.Run("description length is counted in characters", func(t *testing.T) {
		longest := strings.Repeat("\U0001F600", 2000)

		resource.Test(t, resource.TestCase{
			PreCheck:                 testAccPreCheckV2API(t),
			ProtoV6ProviderFactories: testAccProtoV6MuxServerFactory,
			Steps: []resource.TestStep{
				{
					Config:      testAccConfigAPIKeyDescriptionTest(fmt.Sprintf(`description = %q`, longest+"x"), env.ID),
					ExpectError: regexp.MustCompile(`Invalid Attribute Value Length`),
				},
				{
					Config: testAccConfigAPIKeyDescriptionTest(fmt.Sprintf(`description = %q`, longest), env.ID),
					Check:  testAccCheckAPIKeyDescription(t, longest),
				},
			},
		})
	})
}

func TestAcc_APIKeyResource_upgradeFromVersion050_0(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	c := testAccV2Client(t)
	env := testAccEnvironment(ctx, t, c)

	config := testAccConfigIngestAPIKeyTest("test key", "false", env.ID)

	resource.Test(t, resource.TestCase{
		PreCheck: testAccPreCheckV2API(t),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"honeycombio": {
						VersionConstraint: "0.50.0",
						Source:            "honeycombio/honeycombio",
					},
				},
				Config: config,
				Check:  testAccEnsureAPIKeyExists(t, "honeycombio_api_key.test"),
			},
			{
				ProtoV6ProviderFactories: testAccProtoV6MuxServerFactory,
				Config:                   config,
				PlanOnly:                 true,
			},
		},
	})
}

func TestAcc_APIKeyResource_upgradeFromVersion055_2(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	c := testAccV2Client(t)
	env := testAccEnvironment(ctx, t, c)

	config := testAccConfigIngestAPIKeyTest("test key", "false", env.ID)

	resource.Test(t, resource.TestCase{
		PreCheck: testAccPreCheckV2API(t),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"honeycombio": {
						VersionConstraint: "0.55.2",
						Source:            "honeycombio/honeycombio",
					},
				},
				Config: config,
				Check:  testAccEnsureAPIKeyExists(t, "honeycombio_api_key.test"),
			},
			{ // state written without a description must not produce a diff once refreshed
				ProtoV6ProviderFactories: testAccProtoV6MuxServerFactory,
				Config:                   config,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
		},
	})
}

func TestAPIKeyResource_UpgradeStateV0(t *testing.T) {
	t.Parallel()

	const configKey = `{"id":"hcxlk_01","name":"k","description":"","type":"configuration","environment_id":"hcaen_01",
		"disabled":false,"visible_to_members":true,"key":"s","secret":"s",
		"permissions":[{"send_events":true,"create_datasets":false,"manage_queries":false,
		"run_queries":false,"read_service_maps":false,"manage_public_boards":false,
		"manage_private_boards":false,"manage_slos":true,"manage_triggers":false,
		"manage_recipients":false,"manage_markers":false}]}`

	tests := []struct {
		name, state, want string
	}{
		{
			name: "v0.50.0 ingest key backfills false permissions",
			state: `{"id":"hcaik_01","name":"k","type":"ingest","environment_id":"hcaen_01",
				"disabled":false,"key":"hcaik_01s","secret":"s",
				"permissions":[{"create_datasets":true}]}`,
			want: `{"id":"hcaik_01","name":"k","description":"","type":"ingest","environment_id":"hcaen_01",
				"disabled":false,"visible_to_members":false,"key":"hcaik_01s","secret":"s",
				"permissions":[{"send_events":false,"create_datasets":true,"manage_queries":false,
				"run_queries":false,"read_service_maps":false,"manage_public_boards":false,
				"manage_private_boards":false,"manage_slos":false,"manage_triggers":false,
				"manage_recipients":false,"manage_markers":false}]}`,
		},
		{
			name: "v0.50.0 ingest key without permissions",
			state: `{"id":"hcaik_01","name":"k","type":"ingest","environment_id":"hcaen_01",
				"disabled":false,"key":"hcaik_01s","secret":"s","permissions":[]}`,
			want: `{"id":"hcaik_01","name":"k","description":"","type":"ingest","environment_id":"hcaen_01",
				"disabled":false,"visible_to_members":false,"key":"hcaik_01s","secret":"s",
				"permissions":[]}`,
		},
		{
			name:  "v0.51.0+ configuration key only gains an empty description",
			state: strings.Replace(configKey, `"description":"",`, "", 1),
			want:  configKey,
		},
	}

	ctx := context.Background()
	var schemaResp tfresource.SchemaResponse
	(&apiKeyResource{}).Schema(ctx, tfresource.SchemaRequest{}, &schemaResp)
	typ := schemaResp.Schema.Type().TerraformType(ctx)

	server, err := testAccProtoV6ProviderFactory["honeycombio"]()
	require.NoError(t, err)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := server.UpgradeResourceState(ctx, &tfprotov6.UpgradeResourceStateRequest{
				TypeName: "honeycombio_api_key",
				Version:  0,
				RawState: &tfprotov6.RawState{JSON: []byte(tc.state)},
			})
			require.NoError(t, err)
			require.Empty(t, resp.Diagnostics)

			got, err := resp.UpgradedState.Unmarshal(typ)
			require.NoError(t, err)
			want, err := tftypes.ValueFromJSONWithOpts([]byte(tc.want), typ, tftypes.ValueFromJSONOpts{})
			require.NoError(t, err)

			diffs, err := want.Diff(got)
			require.NoError(t, err)
			require.Empty(t, diffs)
		})
	}
}

func testAccConfigIngestAPIKeyTest(name, disabled, envID string) string {
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

func testAccConfigConfigurationAPIKeyTest(name, disabled, visibleToMembers, envID string) string {
	return fmt.Sprintf(`
resource "honeycombio_api_key" "test" {
  name               = "%s"
  type               = "configuration"
  disabled           = %s
  visible_to_members = %s

  environment_id = "%s"

  permissions {
    send_events           = true
	create_datasets       = true
	manage_queries        = true
	run_queries           = true
	read_service_maps     = true
	manage_public_boards  = true
	manage_slos           = true
	manage_triggers       = true
	manage_recipients     = true
	manage_markers        = true
  }
}`, name, disabled, visibleToMembers, envID)
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

func testAccConfigAPIKeyDescriptionTest(descriptionAttr, envID string) string {
	return fmt.Sprintf(`
resource "honeycombio_api_key" "test" {
  name = "test key"
  type = "ingest"
  %s

  environment_id = "%s"

  permissions {
    create_datasets = true
  }
}`, descriptionAttr, envID)
}

// testAccCheckAPIKeyDescription checks the description held by both the
// Terraform state and the API for the "honeycombio_api_key.test" resource.
func testAccCheckAPIKeyDescription(t *testing.T, want string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttr("honeycombio_api_key.test", "description", want),
		func(s *terraform.State) error {
			rs, ok := s.RootModule().Resources["honeycombio_api_key.test"]
			if !ok {
				return errors.New("resource not found: honeycombio_api_key.test")
			}

			key, err := testAccV2Client(t).APIKeys.Get(context.Background(), rs.Primary.ID)
			if err != nil {
				return fmt.Errorf("failed to fetch API key: %w", err)
			}
			if key.Description == nil || *key.Description != want {
				return fmt.Errorf("API reports description %v, want %q", key.Description, want)
			}
			return nil
		},
	)
}
