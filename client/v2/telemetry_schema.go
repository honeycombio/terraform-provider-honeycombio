package v2

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/jsonapi"

	hnyclient "github.com/honeycombio/terraform-provider-honeycombio/client"
)

// Compile-time proof of interface implementation.
var _ TelemetrySchemas = (*telemetrySchemas)(nil)

// TelemetrySchemas manages the custom telemetry schema of the Team.
//
// A Team has at most one schema, so there is no ID to pass to any of the methods.
type TelemetrySchemas interface {
	// Get returns the Team's schema.
	// It returns a DetailedError for which IsNotFound is true when the Team has none.
	Get(ctx context.Context) (*TelemetrySchema, error)
	// Put creates the Team's schema or replaces the existing one.
	Put(ctx context.Context, schema *TelemetrySchema) (*TelemetrySchema, error)
	// Delete removes the Team's schema. It succeeds if the Team has none.
	Delete(ctx context.Context) error
}

const telemetrySchemaPath = "/2/teams/%s/telemetry-schema"

type telemetrySchemas struct {
	client *Client
}

func (t *telemetrySchemas) Get(ctx context.Context) (*TelemetrySchema, error) {
	slug, err := t.client.teamSlug(ctx)
	if err != nil {
		return nil, err
	}
	r, err := t.client.Do(ctx,
		http.MethodGet,
		fmt.Sprintf(telemetrySchemaPath, slug),
		nil,
	)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		return nil, hnyclient.ErrorFromResponse(r)
	}

	schema := new(TelemetrySchema)
	if err := jsonapi.UnmarshalPayload(r.Body, schema); err != nil {
		return nil, err
	}
	return schema, nil
}

func (t *telemetrySchemas) Put(ctx context.Context, s *TelemetrySchema) (*TelemetrySchema, error) {
	slug, err := t.client.teamSlug(ctx)
	if err != nil {
		return nil, err
	}
	r, err := t.client.Do(ctx,
		http.MethodPut,
		fmt.Sprintf(telemetrySchemaPath, slug),
		s,
	)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		return nil, hnyclient.ErrorFromResponse(r)
	}

	schema := new(TelemetrySchema)
	if err := jsonapi.UnmarshalPayload(r.Body, schema); err != nil {
		return nil, err
	}
	return schema, nil
}

func (t *telemetrySchemas) Delete(ctx context.Context) error {
	slug, err := t.client.teamSlug(ctx)
	if err != nil {
		return err
	}
	r, err := t.client.Do(ctx,
		http.MethodDelete,
		fmt.Sprintf(telemetrySchemaPath, slug),
		nil,
	)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusNoContent {
		return hnyclient.ErrorFromResponse(r)
	}
	return nil
}
