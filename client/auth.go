package client

import (
	"context"
)

// The Auth endpoint lists authorizations that have been granted for an API key
// within a team and environment.
//
// API docs: https://docs.honeycomb.io/api/auth/
type Auth interface {
	// List all authorizations for this API key in this team and environment.
	List(ctx context.Context) (AuthMetadata, error)
}

// auth implements Auth.
type auth struct {
	client *Client
}

// Compile-time proof of interface implementation by type auth.
var _ Auth = (*auth)(nil)

type AuthMetadata struct {
	// Authorizations granted to this API key.
	APIKeyAccess struct {
		Boards         bool `json:"boards"`
		Columns        bool `json:"columns"`
		CreateDatasets bool `json:"create_datasets"`
		Events         bool `json:"events"`
		Markers        bool `json:"markers"`
		Queries        bool `json:"queries"`
		Recipients     bool `json:"recipients"`
		SLOs           bool `json:"slos"`
		Triggers       bool `json:"triggers"`
	} `json:"api_key_access"`
	Environment struct {
		// Name is empty for Classic environments.
		Name string `json:"name"`
		// Slug is empty for Classic environments.
		Slug string `json:"slug"`
	} `json:"environment"`
	Team struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	} `json:"team"`
}

func (s *auth) List(ctx context.Context) (AuthMetadata, error) {
	var r AuthMetadata
	err := s.client.Do(ctx, "GET", "/1/auth", nil, &r)
	return r, err
}
