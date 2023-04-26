package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tf5server"
	"github.com/hashicorp/terraform-plugin-mux/tf5muxserver"

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

	// build a pair of V5 Provider Servers to bridge the Plugin SDK based provider
	// and the new Plugin Framework provider as things are migrated
	providers := []func() tfprotov5.ProviderServer{
		providerserver.NewProtocol5(provider.New(providerVersion)),
		// legacy terraform-plugin-sdk provider
		honeycombio.Provider(providerVersion).GRPCProvider,
	}

	muxServer, err := tf5muxserver.NewMuxServer(ctx, providers...)
	if err != nil {
		log.Fatal(err)
	}

	var serveOpts []tf5server.ServeOpt
	if debugMode {
		serveOpts = append(serveOpts, tf5server.WithManagedDebug())
	}

	err = tf5server.Serve(
		"registry.terraform.io/honeycombio/honeycombio",
		muxServer.ProviderServer,
		serveOpts...,
	)
	if err != nil {
		log.Fatal(err)
	}
}
