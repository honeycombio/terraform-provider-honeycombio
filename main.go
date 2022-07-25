package main

import (
	"flag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/honeycombio/terraform-provider-honeycombio/honeycombio"
)

func main() {
	var debugMode bool

	flag.BoolVar(&debugMode, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := &plugin.ServeOpts{
		Debug:        debugMode,
		ProviderAddr: "registry.terraform.io/honeycombio/honeycombio",
		ProviderFunc: honeycombio.Provider,
	}

	plugin.Serve(opts)
}
