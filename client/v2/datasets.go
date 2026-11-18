package v2

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/jsonapi"

	hnyclient "github.com/honeycombio/terraform-provider-honeycombio/client"
)

// Compile-time proof of interface implementation.
var _ Datasets = (*datasets)(nil)

// Datasets reads and changes the datasets in an Environment.
//
// The v2 API can't create or delete datasets; use the v1 client for that.
type Datasets interface {
	// List lists the Environment's datasets. It accepts Filter("slug", slug).
	List(ctx context.Context, environmentID string, opts ...ListOption) (*Pager[Dataset], error)
	Get(ctx context.Context, environmentID, id string) (*Dataset, error)
	// Update changes only the fields of ds that are set.
	Update(ctx context.Context, environmentID string, ds *Dataset) (*Dataset, error)
}

type datasets struct {
	client *Client
}

const (
	datasetsPath     = "/2/teams/%s/environments/%s/datasets"
	datasetsByIDPath = "/2/teams/%s/environments/%s/datasets/%s"
)

func (d *datasets) List(ctx context.Context, environmentID string, opts ...ListOption) (*Pager[Dataset], error) {
	slug, err := d.client.teamSlug(ctx)
	if err != nil {
		return nil, err
	}
	return NewPager[Dataset](d.client, fmt.Sprintf(datasetsPath, slug, environmentID), opts...)
}

func (d *datasets) Get(ctx context.Context, environmentID, id string) (*Dataset, error) {
	return d.do(ctx, http.MethodGet, environmentID, id, nil)
}

func (d *datasets) Update(ctx context.Context, environmentID string, ds *Dataset) (*Dataset, error) {
	return d.do(ctx, http.MethodPatch, environmentID, ds.ID, ds)
}

func (d *datasets) do(ctx context.Context, method, environmentID, id string, body *Dataset) (*Dataset, error) {
	slug, err := d.client.teamSlug(ctx)
	if err != nil {
		return nil, err
	}
	// A nil *Dataset in an interface is not a nil interface, and Do would
	// send it as a body.
	var payload any
	if body != nil {
		payload = body
	}
	r, err := d.client.Do(ctx, method, fmt.Sprintf(datasetsByIDPath, slug, environmentID, id), payload)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		return nil, hnyclient.ErrorFromResponse(r)
	}

	ds := new(Dataset)
	if err := jsonapi.UnmarshalPayload(r.Body, ds); err != nil {
		return nil, err
	}
	return ds, nil
}
