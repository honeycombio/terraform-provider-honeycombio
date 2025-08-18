.PHONY: build testacc lint sweep
default: testacc

build:
	go build -o terraform-provider-honeycombio

testacc:
	TF_ACC=1 go test -v ./...

lint:
# VSCode requires the binary be named golangci-lint-v2, so we check for both names
	@if command -v golangci-lint-v2 >/dev/null 2>&1; then \
		golangci-lint-v2 run; \
	else \
		golangci-lint run; \
	fi

sweep:
# the sweep flag requires a string to be passed, but it is not used
	@echo "WARNING: This will destroy resources. Use only in development teams."
	go test ./internal/provider -v -timeout 5m -sweep=env

