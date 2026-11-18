package v2

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"reflect"

	"github.com/google/go-querystring/query"
	"github.com/hashicorp/jsonapi"

	hnyclient "github.com/honeycombio/terraform-provider-honeycombio/client"
)

type PaginationLinks struct {
	Next *string `json:"next,omitempty"`
}

type ListOptions struct {
	// PageSize is the number of results to return per page.
	// Default is 20, max is 100.
	PageSize int `url:"page[size],omitempty"`

	filters map[string]string
}

type ListOption func(*ListOptions)

const defaultPageSize = 20

func PageSize(size int) ListOption {
	return func(po *ListOptions) { po.PageSize = size }
}

// Filter asks for only the results whose field has value, as the query
// parameter filter[field]. Each list endpoint documents the fields it accepts.
func Filter(field, value string) ListOption {
	return func(po *ListOptions) {
		if po.filters == nil {
			po.filters = make(map[string]string)
		}
		po.filters[field] = value
	}
}

type Pager[T any] struct {
	client *Client
	next   *string
	opts   ListOptions
}

func NewPager[T any](
	c *Client,
	url string,
	os ...ListOption,
) (*Pager[T], error) {
	var opts ListOptions
	for _, o := range os {
		o(&opts)
	}
	if opts.PageSize == 0 {
		opts.PageSize = defaultPageSize
	}

	u, err := c.BaseURL.Parse(url)
	if err != nil {
		return nil, err
	}
	// add any options to the URL
	v, err := query.Values(opts)
	if err != nil {
		return nil, err
	}
	for field, value := range opts.filters {
		v.Set("filter["+field+"]", value)
	}
	u.RawQuery = v.Encode()
	nextUrl := u.RequestURI()

	return &Pager[T]{
		client: c,
		next:   &nextUrl,
		opts:   opts,
	}, nil
}

// HasNext returns true if there are more results to fetch.
func (p *Pager[T]) HasNext() bool { return p.next != nil }

// Next fetches the next page of results.
func (p *Pager[T]) Next(ctx context.Context) ([]*T, error) {
	if p.next == nil {
		return nil, nil
	}
	r, err := p.client.Do(
		ctx,
		http.MethodGet,
		*p.next,
		nil,
	)
	if err != nil {
		return nil, err
	}

	if r.StatusCode != http.StatusOK {
		return nil, hnyclient.ErrorFromResponse(r)
	}
	pagination, err := parsePagination(r)
	if err != nil {
		return nil, err
	}

	payload, err := jsonapi.UnmarshalManyPayload(r.Body, reflect.TypeOf(new(T)))
	if err != nil {
		return nil, err
	}
	items := make([]*T, len(payload))
	for i, obj := range payload {
		if item, ok := obj.(*T); ok {
			items[i] = item
		}
	}

	// update 'next' for the next fetch
	p.next = pagination.Next

	return items, nil
}

func parsePagination(r *http.Response) (*PaginationLinks, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	if err = r.Body.Close(); err != nil {
		return nil, err
	}

	var raw struct {
		PaginationLinks `json:"links"`
	}
	if err := json.NewDecoder(bytes.NewBuffer(body)).Decode(&raw); err != nil {
		return nil, err
	}

	// put body back to be used properly downstream
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	return &raw.PaginationLinks, nil
}
