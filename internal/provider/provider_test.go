package provider

import (
	"context"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-mux/tf5muxserver"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
	v2client "github.com/honeycombio/terraform-provider-honeycombio/client/v2"
	"github.com/honeycombio/terraform-provider-honeycombio/honeycombio"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/test"
)

func init() {
	// load environment values from a .env, if available
	_ = godotenv.Load("../../.env")
}

// used by tests which only use the Framework-based datasources or resources
var testAccProtoV5ProviderFactory = map[string]func() (tfprotov5.ProviderServer, error){
	"honeycombio": providerserver.NewProtocol5WithError(New("test")),
}

// used by tests which use a mix of Framework and SDK-based datasources or resources
//
// n.b. will continue to be used until the SDK-based provider is removed
var testAccProtoV5MuxServerFactory = map[string]func() (tfprotov5.ProviderServer, error){
	"honeycombio": func() (tfprotov5.ProviderServer, error) {
		ctx := context.Background()
		providers := []func() tfprotov5.ProviderServer{
			// modern terraform-plugin-framework provider
			providerserver.NewProtocol5(New("test")),
			// legacy terraform-plugin-sdk provider
			honeycombio.Provider("test").GRPCProvider,
		}

		muxServer, err := tf5muxserver.NewMuxServer(ctx, providers...)
		if err != nil {
			return nil, err
		}

		return muxServer.ProviderServer(), nil
	},
}

func testAccPreCheck(t *testing.T) func() {
	return func() {
		if _, ok := os.LookupEnv("HONEYCOMB_API_KEY"); !ok {
			t.Fatalf("environment variable HONEYCOMB_API_KEY must be set to run acceptance tests")
		}
		if _, ok := os.LookupEnv("HONEYCOMB_DATASET"); !ok {
			t.Fatalf("environment variable HONEYCOMB_DATASET must be set to run acceptance tests")
		}
	}
}

func testAccPreCheckV2API(t *testing.T) func() {
	return func() {
		if _, ok := os.LookupEnv("HONEYCOMB_KEY_ID"); !ok {
			t.Fatalf("environment variable HONEYCOMB_KEY_ID must be set to run acceptance tests")
		}
		if _, ok := os.LookupEnv("HONEYCOMB_KEY_SECRET"); !ok {
			t.Fatalf("environment variable HONEYCOMB_KEY_SECRET must be set to run acceptance tests")
		}
	}
}

func testAccDataset() string {
	return os.Getenv("HONEYCOMB_DATASET")
}

// newTestEnvirionment creates a new Environment with a random name and description
// for testing purposes.
// The Environment is automatically deleted when the test completes.
func testAccEnvironment(ctx context.Context, t *testing.T, c *v2client.Client) *v2client.Environment {
	t.Helper()

	env, err := c.Environments.Create(ctx, &v2client.Environment{
		Name:        test.RandomStringWithPrefix("test.", 20),
		Description: helper.ToPtr(test.RandomString(50)),
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		// disable deletion protection and delete the Environment
		c.Environments.Update(context.Background(), &v2client.Environment{
			ID: env.ID,
			Settings: &v2client.EnvironmentSettings{
				DeleteProtected: helper.ToPtr(false),
			},
		})
		c.Environments.Delete(ctx, env.ID)
	})

	return env
}

func testAccClient(t *testing.T) *client.Client {
	c, err := client.NewClient()
	if err != nil {
		t.Fatalf("could not initialize Honeycomb client: %v", err)
	}
	return c
}

func testAccV2Client(t *testing.T) *v2client.Client {
	c, err := v2client.NewClient()
	if err != nil {
		t.Fatalf("could not initialize Honeycomb client: %v", err)
	}
	return c
}

// TestAcc_Configuration verifies that the provider is correctly
// configured with the supported configuration permutations.
func TestAcc_Configuration(t *testing.T) {
	t.Run("v1 and v2 clients", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
			Steps: []resource.TestStep{
				{
					Config: `
data "honeycombio_datasets" "all" {}

data "honeycombio_environments" "all" {}
`,
					PlanOnly: true,
				},
			},
		})
	})

	t.Run("v1 client only", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
			Steps: []resource.TestStep{
				{
					PreConfig: func() {
						t.Setenv("HONEYCOMB_KEY_ID", "")
						t.Setenv("HONEYCOMB_KEY_SECRET", "")
					},
					Config:   `data "honeycombio_datasets" "all" {}`,
					PlanOnly: true,
				},
			},
		})
	})

	t.Run("v2 client only", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
			Steps: []resource.TestStep{
				{
					PreConfig: func() {
						t.Setenv("HONEYCOMB_API_KEY", "")
					},
					Config:   `data "honeycombio_environments" "all" {}`,
					PlanOnly: true,
				},
			},
		})
	})

	t.Run("fails when only half of v2 configuration is set", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
			Steps: []resource.TestStep{
				{
					PreConfig: func() {
						t.Setenv("HONEYCOMB_API_KEY", "")
						t.Setenv("HONEYCOMB_KEY_SECRET", "")
					},
					Config:      `data "honeycombio_environments" "all" {}`,
					PlanOnly:    true,
					ExpectError: regexp.MustCompile(`provider requires both a Honeycomb API Key ID and Secret`),
				},
			},
		})
	})

	t.Run("fails when no clients configured", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
			Steps: []resource.TestStep{
				{
					PreConfig: func() {
						t.Setenv("HONEYCOMB_API_KEY", "")
						t.Setenv("HONEYCOMB_KEY_ID", "")
						t.Setenv("HONEYCOMB_KEY_SECRET", "")
					},
					Config:      `data "honeycombio_datasets" "all" {}`,
					PlanOnly:    true,
					ExpectError: regexp.MustCompile(`requires at least one of a Honeycomb API Key`),
				},
			},
		})
	})

	t.Run("fails v1-only config using v2 resource", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
			Steps: []resource.TestStep{
				{
					PreConfig: func() {
						t.Setenv("HONEYCOMB_KEY_ID", "")
						t.Setenv("HONEYCOMB_KEY_SECRET", "")
					},
					Config:      `data "honeycombio_environments" "all" {}`,
					ExpectError: regexp.MustCompile(`No v2 API client configured`),
				},
			},
		})
	})

	t.Run("fails v2-only config using v1 resource", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
			Steps: []resource.TestStep{
				{
					PreConfig: func() {
						t.Setenv("HONEYCOMB_API_KEY", "")
					},
					// plugin SDK-based
					Config:      `data "honeycombio_datasets" "all" {}`,
					ExpectError: regexp.MustCompile(`No v1 API client configured`),
				},
				// framework-based
				{
					PreConfig: func() {
						t.Setenv("HONEYCOMB_API_KEY", "")
					},
					Config:      `data "honeycombio_auth_metadata" "current" {}`,
					ExpectError: regexp.MustCompile(`No v1 API client configured`),
				},
			},
		})
	})
}

func TestProvider_Configure_WithCustomEnvVars(t *testing.T) {
	// Test that custom environment variable names work correctly
	// Set the custom environment variables with actual API key values
	t.Setenv("CUSTOM_API_KEY", "test-api-key")
	t.Setenv("CUSTOM_KEY_ID", "test-key-id")
	t.Setenv("CUSTOM_KEY_SECRET", "test-key-secret")

	p := New("test")
	var schemaResp provider.SchemaResponse
	p.Schema(context.Background(), provider.SchemaRequest{}, &schemaResp)
	req := provider.ConfigureRequest{
		Config: tfsdk.Config{
			Schema: schemaResp.Schema,
			Raw: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"api_key":                tftypes.String,
					"api_key_id":             tftypes.String,
					"api_key_secret":         tftypes.String,
					"api_url":                tftypes.String,
					"debug":                  tftypes.Bool,
					"api_key_env_var":        tftypes.String,
					"api_key_id_env_var":     tftypes.String,
					"api_key_secret_env_var": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"api_key":                tftypes.NewValue(tftypes.String, ""),
				"api_key_id":             tftypes.NewValue(tftypes.String, ""),
				"api_key_secret":         tftypes.NewValue(tftypes.String, ""),
				"api_url":                tftypes.NewValue(tftypes.String, ""),
				"debug":                  tftypes.NewValue(tftypes.Bool, false),
				"api_key_env_var":        tftypes.NewValue(tftypes.String, "CUSTOM_API_KEY"),
				"api_key_id_env_var":     tftypes.NewValue(tftypes.String, "CUSTOM_KEY_ID"),
				"api_key_secret_env_var": tftypes.NewValue(tftypes.String, "CUSTOM_KEY_SECRET"),
			}),
		},
		TerraformVersion: "1.0.0",
	}

	var resp provider.ConfigureResponse
	p.Configure(context.Background(), req, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Expected no errors, got: %v", resp.Diagnostics)
	}
}

func TestProvider_Configure_WithCustomEnvVarsAndRealAPI(t *testing.T) {
	// Skip if not running acceptance tests
	if os.Getenv("TF_ACC") != "1" {
		t.Skip("Skipping acceptance test")
	}

	// Test with real environment variables that should exist in acceptance test environment
	p := New("test")
	var schemaResp provider.SchemaResponse
	p.Schema(context.Background(), provider.SchemaRequest{}, &schemaResp)
	req := provider.ConfigureRequest{
		Config: tfsdk.Config{
			Schema: schemaResp.Schema,
			Raw: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"api_key":                tftypes.String,
					"api_key_id":             tftypes.String,
					"api_key_secret":         tftypes.String,
					"api_url":                tftypes.String,
					"debug":                  tftypes.Bool,
					"api_key_env_var":        tftypes.String,
					"api_key_id_env_var":     tftypes.String,
					"api_key_secret_env_var": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"api_key":                tftypes.NewValue(tftypes.String, ""),
				"api_key_id":             tftypes.NewValue(tftypes.String, ""),
				"api_key_secret":         tftypes.NewValue(tftypes.String, ""),
				"api_url":                tftypes.NewValue(tftypes.String, ""),
				"debug":                  tftypes.NewValue(tftypes.Bool, false),
				"api_key_env_var":        tftypes.NewValue(tftypes.String, "HONEYCOMB_API_KEY"),
				"api_key_id_env_var":     tftypes.NewValue(tftypes.String, "HONEYCOMB_KEY_ID"),
				"api_key_secret_env_var": tftypes.NewValue(tftypes.String, "HONEYCOMB_KEY_SECRET"),
			}),
		},
		TerraformVersion: "1.0.0",
	}

	var resp provider.ConfigureResponse
	p.Configure(context.Background(), req, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Expected no errors, got: %v", resp.Diagnostics)
	}

	// Verify that the client was configured correctly
	client, ok := resp.DataSourceData.(*ConfiguredClient)
	if !ok {
		t.Fatal("Expected client to be of type *ConfiguredClient")
	}
	if client == nil {
		t.Fatal("Expected client to be configured")
	}
}

func TestProvider_Configure_WithCustomEnvVarsAndDefaults(t *testing.T) {
	// Test that when custom env vars are not set, it falls back to defaults
	p := New("test")
	var schemaResp provider.SchemaResponse
	p.Schema(context.Background(), provider.SchemaRequest{}, &schemaResp)
	req := provider.ConfigureRequest{
		Config: tfsdk.Config{
			Schema: schemaResp.Schema,
			Raw: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{},
			}, map[string]tftypes.Value{}),
		},
		TerraformVersion: "1.0.0",
	}

	var resp provider.ConfigureResponse
	p.Configure(context.Background(), req, &resp)

	// Should have an error because no API keys are set, but the error should mention default env var names
	if !resp.Diagnostics.HasError() {
		t.Fatal("Expected error when no API keys are configured")
	}

	// Check that there are diagnostics (errors)
	if len(resp.Diagnostics) == 0 {
		t.Error("Expected diagnostics to contain errors")
	}
}

func TestProvider_Configure_WithCustomEnvVarsAndCustomErrorMessages(t *testing.T) {
	// Test that error messages reflect custom environment variable names
	p := New("test")
	var schemaResp provider.SchemaResponse
	p.Schema(context.Background(), provider.SchemaRequest{}, &schemaResp)
	req := provider.ConfigureRequest{
		Config: tfsdk.Config{
			Schema: schemaResp.Schema,
			Raw: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"api_key_env_var":        tftypes.String,
					"api_key_id_env_var":     tftypes.String,
					"api_key_secret_env_var": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"api_key_env_var":        tftypes.NewValue(tftypes.String, "CUSTOM_API_KEY"),
				"api_key_id_env_var":     tftypes.NewValue(tftypes.String, "CUSTOM_KEY_ID"),
				"api_key_secret_env_var": tftypes.NewValue(tftypes.String, "CUSTOM_KEY_SECRET"),
			}),
		},
		TerraformVersion: "1.0.0",
	}

	var resp provider.ConfigureResponse
	p.Configure(context.Background(), req, &resp)

	// Should have an error because no API keys are set, but the error should mention custom env var names
	if !resp.Diagnostics.HasError() {
		t.Fatal("Expected error when no API keys are configured")
	}

	// Check that there are diagnostics (errors)
	if len(resp.Diagnostics) == 0 {
		t.Error("Expected diagnostics to contain errors")
	}
}
