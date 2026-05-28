package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6/tf6server"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"

	"github.com/honeycombio/terraform-provider-honeycombio/honeycombio"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/provider"
)

// providerVersion represents the current version of the provider. It should be
// overwritten during the release process.
var providerVersion = "dev"

func main() {
	ctx := context.Background()

	var debugMode bool
	flag.BoolVar(&debugMode, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	// build a pair of V6 Provider Servers to bridge the Plugin SDK based provider
	// and the new Plugin Framework provider as things are migrated. We use the upgrade server
	// as we've not fully upgraded to the framework server yet when we were using v5, so there might
	// still be some risk of version incompatibilities.
	providers := []func() tfprotov6.ProviderServer{
		func() tfprotov6.ProviderServer {
			upgradedFrameworkServer, err := tf5to6server.UpgradeServer(
				context.Background(),
				providerserver.NewProtocol5(provider.New(providerVersion)),
			)
			if err != nil {
				log.Fatal(err)
			}

			return upgradedFrameworkServer
		},
		func() tfprotov6.ProviderServer {
			upgradedSDKServer, err := tf5to6server.UpgradeServer(
				context.Background(),
				honeycombio.Provider(providerVersion).GRPCProvider,
			)
			if err != nil {
				log.Fatal(err)
			}

			return upgradedSDKServer
		},
	}

	muxServer, err := tf6muxserver.NewMuxServer(ctx, providers...)
	if err != nil {
		log.Fatal(err)
	}

	var serveOpts []tf6server.ServeOpt
	if debugMode {
		serveOpts = append(serveOpts, tf6server.WithManagedDebug())
	}

	err = tf6server.Serve(
		"registry.terraform.io/honeycombio/honeycombio",
		muxServer.ProviderServer,
		serveOpts...,
	)
	if err != nil {
		log.Fatal(err)
	}
}
