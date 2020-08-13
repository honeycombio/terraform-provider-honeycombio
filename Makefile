build:
	go build -o terraform-provider-honeycombio

testacc:
	TF_ACC=1 go test -v ./...

# Terraform 0.13 only: build the repository and install the provider in one of
# the local mirror directories following the new fileystem layout. Additionally,
# we have to specify a version.
#
# https://www.terraform.io/docs/commands/cli-config.html#implied-local-mirror-directories
# https://www.terraform.io/upgrade-guides/0-13.html#new-filesystem-layout-for-local-copies-of-providers

version = 99.0.0
os_arch = $(shell go version | cut -d' ' -f4 | tr / _)
provider_path = registry.terraform.io/kvrhdn/honeycombio/$(version)/$(os_arch)/

install_macos:
	go build -o terraform-provider-honeycombio_$(version)

	mkdir -p ~/Library/Application\ Support/io.terraform/plugins/$(provider_path)
	cp terraform-provider-honeycombio_$(version)  ~/Library/Application\ Support/io.terraform/plugins/$(provider_path)

uninstall_macos:
	rm -r ~/Library/Application\ Support/io.terraform/plugins/registry.terraform.io/kvrhdn

.PHONY: build testacc install
