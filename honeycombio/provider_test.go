package honeycombio

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"
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
var testAccProtoV6ProviderFactory = map[string]func() (tfprotov6.ProviderServer, error){
	"honeycombio": func() (tfprotov6.ProviderServer, error) {
		ctx := context.Background()
		providers := []func() tfprotov6.ProviderServer{
			providerserver.NewProtocol6(frameworkProvider.New("test")),
			func() tfprotov6.ProviderServer {
				upgradedSDKServer, err := tf5to6server.UpgradeServer(
					ctx,
					Provider("test").GRPCProvider,
				)
				if err != nil {
					log.Fatal(err)
				}

				return upgradedSDKServer
			},
		}

		muxServer, err := tf6muxserver.NewMuxServer(ctx, providers...)
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
