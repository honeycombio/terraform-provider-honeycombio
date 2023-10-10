package provider

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-mux/tf5muxserver"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/joho/godotenv"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/honeycombio"
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

func testAccDataset() string {
	return os.Getenv("HONEYCOMB_DATASET")
}

func testAccClient(t *testing.T) *client.Client {
	c, err := client.NewClient(client.DefaultConfig())
	if err != nil {
		t.Fatalf("could not initialize Honeycomb client: %v", err)
	}
	return c
}

// TestMuxServer verifies that a V5 Mux Server can be properly created while
// the Plugin SDK and the Plugin Framework are both in use in the provider
func TestAcc_MuxServer(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
		Steps: []resource.TestStep{
			{
				// simple smoketest by accessing a datasource
				Config: `data "honeycombio_datasets" "all" {}`,
			},
		},
	})
}
