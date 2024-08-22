package v2

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/jsonapi"

	hnyclient "github.com/honeycombio/terraform-provider-honeycombio/client"
)

// Compile-time proof of interface implementation.
var _ APIKeys = (*apiKeys)(nil)

type APIKeys interface {
	Create(ctx context.Context, key *APIKey) (*APIKey, error)
	Get(ctx context.Context, id string) (*APIKey, error)
	Update(ctx context.Context, key *APIKey) (*APIKey, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, opts ...ListOption) (*Pager[APIKey], error)
}

const (
	apiKeysPath     = "/2/teams/%s/api-keys"
	apiKeysByIDPath = "/2/teams/%s/api-keys/%s"
)

type apiKeys struct {
	client   *Client
	authinfo *AuthMetadata
}

func (a *apiKeys) Create(ctx context.Context, k *APIKey) (*APIKey, error) {
	r, err := a.client.Do(ctx,
		http.MethodPost,
		fmt.Sprintf(apiKeysPath, a.authinfo.Team.Slug),
		k,
	)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusCreated {
		return nil, hnyclient.ErrorFromResponse(r)
	}

	key := new(APIKey)
	if err := jsonapi.UnmarshalPayload(r.Body, key); err != nil {
		return nil, err
	}
	return key, nil
}

func (a *apiKeys) Get(ctx context.Context, id string) (*APIKey, error) {
	r, err := a.client.Do(ctx,
		http.MethodGet,
		fmt.Sprintf(apiKeysByIDPath, a.authinfo.Team.Slug, id),
		nil,
	)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		return nil, hnyclient.ErrorFromResponse(r)
	}

	key := new(APIKey)
	if err := jsonapi.UnmarshalPayload(r.Body, key); err != nil {
		return nil, err
	}
	return key, nil
}

func (a *apiKeys) Update(ctx context.Context, k *APIKey) (*APIKey, error) {
	r, err := a.client.Do(ctx,
		http.MethodPatch,
		fmt.Sprintf(apiKeysByIDPath, a.authinfo.Team.Slug, k.ID),
		k,
	)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		return nil, hnyclient.ErrorFromResponse(r)
	}

	key := new(APIKey)
	if err := jsonapi.UnmarshalPayload(r.Body, key); err != nil {
		return nil, err
	}
	return key, nil
}

func (a *apiKeys) Delete(ctx context.Context, id string) error {
	r, err := a.client.Do(ctx,
		http.MethodDelete,
		fmt.Sprintf(apiKeysByIDPath, a.authinfo.Team.Slug, id),
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

func (a *apiKeys) List(ctx context.Context, os ...ListOption) (*Pager[APIKey], error) {
	return NewPager[APIKey](
		a.client,
		fmt.Sprintf(apiKeysPath, a.authinfo.Team.Slug),
		os...,
	)
}
