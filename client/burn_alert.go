package client

import (
	"context"
	"fmt"
	"time"
)

// BurnAlerts describe all the BurnAlert-related methods that the Honeycomb API supports.
//
// API docs: https://docs.honeycomb.io/api/burn-alerts/
type BurnAlerts interface {
	// List all BurnAlerts associated with a SLO. Returns ErrNotFound if there is no SLO with
	// the given ID in this dataset.
	ListForSLO(ctx context.Context, dataset string, sloId string) ([]BurnAlert, error)

	// Get a BurnAlert by its ID. Returns ErrNotFound if there is no BurnAlert with
	// the given ID in this dataset.
	Get(ctx context.Context, dataset string, id string) (*BurnAlert, error)

	// Create a new BurnAlert on a SLO. When creating a BurnAlert ID may not
	// be set.
	Create(ctx context.Context, dataset string, s *BurnAlert) (*BurnAlert, error)

	// Update an existing BurnAlert.
	Update(ctx context.Context, dataset string, s *BurnAlert) (*BurnAlert, error)

	// Delete a BurnAlert from a dataset.
	Delete(ctx context.Context, dataset string, id string) error
}

// burnalerts implements BurnAlerts.
type burnalerts struct {
	client *Client
}

// Compile-time proof of interface implementation by type burnalerts.
var _ BurnAlerts = (*burnalerts)(nil)

type SLORef struct {
	ID string `json:"id"`
}

type BurnAlert struct {
	ID                string                  `json:"id,omitempty"`
	ExhaustionMinutes int                     `json:"exhaustion_minutes"`
	SLO               SLORef                  `json:"slo"`
	CreatedAt         time.Time               `json:"created_at,omitempty"`
	UpdatedAt         time.Time               `json:"updated_at,omitempty"`
	Recipients        []NotificationRecipient `json:"recipients,omitempty"`
}

func (s *burnalerts) ListForSLO(ctx context.Context, dataset string, sloId string) ([]BurnAlert, error) {
	var b []BurnAlert
	err := s.client.performRequest(ctx, "GET", fmt.Sprintf("/1/burn_alerts/%s?slo_id=%s", urlEncodeDataset(dataset), sloId), nil, &b)
	return b, err
}

func (s *burnalerts) Get(ctx context.Context, dataset string, id string) (*BurnAlert, error) {
	var b BurnAlert
	err := s.client.performRequest(ctx, "GET", fmt.Sprintf("/1/burn_alerts/%s/%s", urlEncodeDataset(dataset), id), nil, &b)
	return &b, err
}

func (s *burnalerts) Create(ctx context.Context, dataset string, data *BurnAlert) (*BurnAlert, error) {
	var b BurnAlert
	err := s.client.performRequest(ctx, "POST", fmt.Sprintf("/1/burn_alerts/%s", urlEncodeDataset(dataset)), data, &b)
	return &b, err
}

func (s *burnalerts) Update(ctx context.Context, dataset string, data *BurnAlert) (*BurnAlert, error) {
	var b BurnAlert
	err := s.client.performRequest(ctx, "PUT", fmt.Sprintf("/1/burn_alerts/%s/%s", urlEncodeDataset(dataset), data.ID), data, &b)
	return &b, err
}

func (s *burnalerts) Delete(ctx context.Context, dataset string, id string) error {
	return s.client.performRequest(ctx, "DELETE", fmt.Sprintf("/1/burn_alerts/%s/%s", urlEncodeDataset(dataset), id), nil, nil)
}
