package v2

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/jsonapi"

	hnyclient "github.com/honeycombio/terraform-provider-honeycombio/client"
)

// Compile-time proof of interface implementation.
var _ Environments = (*environments)(nil)

type Environments interface {
	Create(ctx context.Context, env *Environment) (*Environment, error)
	Get(ctx context.Context, id string) (*Environment, error)
	Update(ctx context.Context, env *Environment) (*Environment, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, opts ...ListOption) (*Pager[Environment], error)
}

type environments struct {
	client   *Client
	authinfo *AuthMetadata
}

const (
	environmentsPath     = "/2/teams/%s/environments"
	environmentsByIDPath = "/2/teams/%s/environments/%s"
)

const (
	EnvironmentColorBlue        = "blue"
	EnvironmentColorGreen       = "green"
	EnvironmentColorGold        = "gold"
	EnvironmentColorRed         = "red"
	EnvironmentColorPurple      = "purple"
	EnvironmentColorLightBlue   = "lightBlue"
	EnvironmentColorLightGreen  = "lightGreen"
	EnvironmentColorLightGold   = "lightGold"
	EnvironmentColorLightRed    = "lightRed"
	EnvironmentColorLightPurple = "lightPurple"
)

func EnvironmentColorTypes() []string {
	return []string{
		EnvironmentColorBlue,
		EnvironmentColorGreen,
		EnvironmentColorGold,
		EnvironmentColorRed,
		EnvironmentColorPurple,
		EnvironmentColorLightBlue,
		EnvironmentColorLightGreen,
		EnvironmentColorLightGold,
		EnvironmentColorLightRed,
		EnvironmentColorLightPurple,
	}
}

func (e *environments) Create(ctx context.Context, env *Environment) (*Environment, error) {
	r, err := e.client.Do(ctx,
		http.MethodPost,
		fmt.Sprintf(environmentsPath, e.authinfo.Team.Slug),
		env,
	)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusCreated {
		return nil, hnyclient.ErrorFromResponse(r)
	}

	environ := new(Environment)
	if err := jsonapi.UnmarshalPayload(r.Body, environ); err != nil {
		return nil, err
	}
	return environ, nil
}

func (e *environments) Get(ctx context.Context, id string) (*Environment, error) {
	r, err := e.client.Do(ctx,
		http.MethodGet,
		fmt.Sprintf(environmentsByIDPath, e.authinfo.Team.Slug, id),
		nil,
	)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		return nil, hnyclient.ErrorFromResponse(r)
	}

	environ := new(Environment)
	if err := jsonapi.UnmarshalPayload(r.Body, environ); err != nil {
		return nil, err
	}
	return environ, nil
}

func (e *environments) Update(ctx context.Context, env *Environment) (*Environment, error) {
	r, err := e.client.Do(ctx,
		http.MethodPatch,
		fmt.Sprintf(environmentsByIDPath, e.authinfo.Team.Slug, env.ID),
		env,
	)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		return nil, hnyclient.ErrorFromResponse(r)
	}

	environ := new(Environment)
	if err := jsonapi.UnmarshalPayload(r.Body, environ); err != nil {
		return nil, err
	}
	return environ, nil
}

func (e *environments) Delete(ctx context.Context, id string) error {
	r, err := e.client.Do(ctx,
		http.MethodDelete,
		fmt.Sprintf(environmentsByIDPath, e.authinfo.Team.Slug, id),
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

func (e *environments) List(ctx context.Context, os ...ListOption) (*Pager[Environment], error) {
	return NewPager[Environment](
		e.client,
		fmt.Sprintf(environmentsPath, e.authinfo.Team.Slug),
		os...,
	)
}
