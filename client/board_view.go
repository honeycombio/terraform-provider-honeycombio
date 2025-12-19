package client

import (
	"context"
	"fmt"
)

// BoardViews describes all the board view-related methods that the Honeycomb API
// supports.
type BoardViews interface {
	// List all views for a board.
	List(ctx context.Context, boardID string) ([]BoardView, error)

	// Get a board view by its ID.
	Get(ctx context.Context, boardID, viewID string) (*BoardView, error)

	// Create a new board view. When creating a new view ID may not be set.
	Create(ctx context.Context, boardID string, view *BoardView) (*BoardView, error)

	// Update an existing board view.
	Update(ctx context.Context, boardID string, view *BoardView) (*BoardView, error)

	// Delete a board view.
	Delete(ctx context.Context, boardID, viewID string) error
}

// boardViews implements BoardViews.
type boardViews struct {
	client *Client
}

// Compile-time proof of interface implementation by type boardViews.
var _ BoardViews = (*boardViews)(nil)

// BoardView represents a Honeycomb board view.
type BoardView struct {
	ID      string            `json:"id,omitempty"`
	Name    string            `json:"name"`
	Filters []BoardViewFilter `json:"filters"`
}

// BoardViewFilter represents a filter within a board view.
// Note: This uses "operation" as the JSON field name, unlike FilterSpec which uses "op".
type BoardViewFilter struct {
	Column    string `json:"column"`
	Operation string `json:"operation"`
	// Value to use with the filter operation. The type of the filter value
	// depends on the operator:
	//  - 'exists' and 'does-not-exist': value should be nil
	//  - 'in' and 'not-in': value should be a []any
	//  - all other ops: value could be a string, int, bool or float
	Value any `json:"value,omitempty"`
}

func (s *boardViews) List(ctx context.Context, boardID string) ([]BoardView, error) {
	var views []BoardView
	err := s.client.Do(ctx, "GET", fmt.Sprintf("/1/boards/%s/views", boardID), nil, &views)
	return views, err
}

func (s *boardViews) Get(ctx context.Context, boardID, viewID string) (*BoardView, error) {
	var view BoardView
	err := s.client.Do(ctx, "GET", fmt.Sprintf("/1/boards/%s/views/%s", boardID, viewID), nil, &view)
	return &view, err
}

func (s *boardViews) Create(ctx context.Context, boardID string, data *BoardView) (*BoardView, error) {
	var view BoardView
	err := s.client.Do(ctx, "POST", fmt.Sprintf("/1/boards/%s/views", boardID), data, &view)
	return &view, err
}

func (s *boardViews) Update(ctx context.Context, boardID string, data *BoardView) (*BoardView, error) {
	var view BoardView
	err := s.client.Do(ctx, "PUT", fmt.Sprintf("/1/boards/%s/views/%s", boardID, data.ID), data, &view)
	return &view, err
}

func (s *boardViews) Delete(ctx context.Context, boardID, viewID string) error {
	return s.client.Do(ctx, "DELETE", fmt.Sprintf("/1/boards/%s/views/%s", boardID, viewID), nil, nil)
}
