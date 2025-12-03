.PHONY: build testacc lint sweep docs docs-check
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

docs:
	@echo "Generating documentation..."
	@go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --provider-dir . --provider-name honeycombio

docs-check:
	@echo "Checking if documentation is up to date..."
	@go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --provider-dir . --provider-name honeycombio
	@git diff --exit-code docs/ || (echo "Documentation is out of date. Run 'make docs' to regenerate." && exit 1)
