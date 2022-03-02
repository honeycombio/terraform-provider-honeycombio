package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/honeycombio/terraform-provider-honeycombio/honeycombio"
)

func main() {
	var debugMode bool

	flag.BoolVar(&debugMode, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := &plugin.ServeOpts{ProviderFunc: honeycombio.Provider}

	if debugMode {
		err := plugin.Debug(context.Background(), "registry.terraform.io/honeycombio/honeycombio", opts)
		if err != nil {
			log.Fatalf(err.Error())
		}
		return
	}

	plugin.Serve(opts)
}
