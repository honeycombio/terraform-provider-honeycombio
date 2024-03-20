package honeycombio

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-mux/tf5muxserver"
	"github.com/joho/godotenv"

	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
	frameworkProvider "github.com/honeycombio/terraform-provider-honeycombio/internal/provider"
)

func init() {
	// load environment values from a .env, if available
	_ = godotenv.Load("../.env")
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

// used by tests which use a mix of Framework and SDK-based datasources or resources
//
// n.b. will continue to be used until the SDK-based provider is removed
var testAccProtoV5ProviderFactory = map[string]func() (tfprotov5.ProviderServer, error){
	"honeycombio": func() (tfprotov5.ProviderServer, error) {
		ctx := context.Background()
		providers := []func() tfprotov5.ProviderServer{
			// modern terraform-plugin-framework provider
			providerserver.NewProtocol5(frameworkProvider.New("test")),
			// legacy terraform-plugin-sdk provider
			Provider("test").GRPCProvider,
		}

		muxServer, err := tf5muxserver.NewMuxServer(ctx, providers...)
		if err != nil {
			return nil, err
		}

		return muxServer.ProviderServer(), nil
	},
}

func testAccClient(t *testing.T) *honeycombio.Client {
	c, err := honeycombio.NewClient()
	if err != nil {
		t.Fatalf("could not initialize honeycombio.Client: %v", err)
	}
	return c
}

func testAccDataset() string {
	return os.Getenv("HONEYCOMB_DATASET")
}
