build:
	go build -o terraform-provider-honeycombio

testacc:
	TF_ACC=1 go test -v ./...

.PHONY: testacc
