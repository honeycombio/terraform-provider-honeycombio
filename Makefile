.PHONY: build testacc lint lint-version sweep docs docs-check
default: testacc

GOLANGCI_LINT_VERSION := v2.9.0

build:
	go build -o terraform-provider-honeycombio

testacc:
	TF_ACC=1 go test -v ./...

lint:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION) run

lint-version:
	@echo $(GOLANGCI_LINT_VERSION)

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
