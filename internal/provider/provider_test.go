package provider

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-mux/tf5muxserver"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/honeycombio"
	"github.com/joho/godotenv"
)

func init() {
	// load environment values from a .env, if available
	_ = godotenv.Load("../../.env")
}

var testAccProtoV5ProviderFactories = map[string]func() (tfprotov5.ProviderServer, error){
	"honeycombio": providerserver.NewProtocol5WithError(New("test")),
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
	cfg := &client.Config{
		APIKey: os.Getenv("HONEYCOMB_API_KEY"),
	}
	c, err := client.NewClient(cfg)
	if err != nil {
		t.Fatalf("could not initialize Honeycomb client: %v", err)
	}
	return c
}

// TestMuxServer verifies that a V5 Mux Server can be properly created while
// the Plugin SDK and the Plugin Framework are both in use in the provider
func TestMuxServer(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: map[string]func() (tfprotov5.ProviderServer, error){
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
		},
		Steps: []resource.TestStep{
			{
				// simple smoketest by accessing a datasource
				Config: `data "honeycombio_datasets" "all" {}`,
			},
		},
	})
}
