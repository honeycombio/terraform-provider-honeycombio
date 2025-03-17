package client

import (
	"context"
	"fmt"
	"time"
)

// SLOs describe all the SLO-related methods that the Honeycomb API supports.
//
// API docs: https://docs.honeycomb.io/api/slos/
type SLOs interface {
	// List all SLOs present in this dataset.
	List(ctx context.Context, dataset string) ([]SLO, error)

	// Get a SLO by its ID.
	Get(ctx context.Context, dataset string, id string) (*SLO, error)

	// Create a new SLO in this dataset. When creating a SLO ID may not
	// be set.
	Create(ctx context.Context, dataset string, s *SLO) (*SLO, error)

	// Update an existing SLO.
	Update(ctx context.Context, dataset string, s *SLO) (*SLO, error)

	// Delete a SLO from the dataset.
	Delete(ctx context.Context, dataset string, id string) error
}

// slos implements SLOs.
type slos struct {
	client *Client
}

// Compile-time proof of interface implementation by type slos.
var _ SLOs = (*slos)(nil)

type SLIRef struct {
	Alias string `json:"alias"`
}

type SLO struct {
	ID               string    `json:"id,omitempty"`
	Name             string    `json:"name"`
	Description      string    `json:"description,omitempty"`
	TimePeriodDays   int       `json:"time_period_days"`
	TargetPerMillion int       `json:"target_per_million"`
	DatasetSlugs     []string  `json:"dataset_slugs,omitempty"`
	SLI              SLIRef    `json:"sli"`
	CreatedAt        time.Time `json:"created_at,omitempty"`
	UpdatedAt        time.Time `json:"updated_at,omitempty"`
}

const Dataset_All = "__all__"

func (s *slos) List(ctx context.Context, dataset string) ([]SLO, error) {
	var r []SLO
	err := s.client.Do(ctx, "GET", fmt.Sprintf("/1/slos/%s", urlEncodeDataset(dataset)), nil, &r)
	return r, err
}

func (s *slos) Get(ctx context.Context, dataset string, id string) (*SLO, error) {
	var r SLO
	err := s.client.Do(ctx, "GET", fmt.Sprintf("/1/slos/%s/%s", urlEncodeDataset(dataset), id), nil, &r)
	return &r, err
}

func (s *slos) Create(ctx context.Context, dataset string, data *SLO) (*SLO, error) {
	var r SLO
	err := s.client.Do(ctx, "POST", fmt.Sprintf("/1/slos/%s", urlEncodeDataset(dataset)), data, &r)
	return &r, err
}

func (s *slos) Update(ctx context.Context, dataset string, data *SLO) (*SLO, error) {
	var r SLO
	err := s.client.Do(ctx, "PUT", fmt.Sprintf("/1/slos/%s/%s", urlEncodeDataset(dataset), data.ID), data, &r)
	return &r, err
}

func (s *slos) Delete(ctx context.Context, dataset string, id string) error {
	return s.client.Do(ctx, "DELETE", fmt.Sprintf("/1/slos/%s/%s", urlEncodeDataset(dataset), id), nil, nil)
}
