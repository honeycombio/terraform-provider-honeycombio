.PHONY: build testacc lint sweep
default: testacc

build:
	go build -o terraform-provider-honeycombio

testacc:
	TF_ACC=1 go test -v ./...

lint:
	golangci-lint run

sweep:
# the sweep flag requires a string to be passed, but it is not used
	@echo "WARNING: This will destroy resources. Use only in development teams."
	go test ./internal/provider -v -timeout 5m -sweep=env

