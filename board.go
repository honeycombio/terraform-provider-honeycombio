package honeycombio

import (
	"context"
	"fmt"
)

// Boards describes all the board-related methods that the Honeycomb API
// supports.
//
// API docs: https://docs.honeycomb.io/api/boards-api/
type Boards interface {
	// List all boards.
	List(ctx context.Context) ([]Board, error)

	// Get a board by its ID. Returns ErrNotFound if there is no board with the
	// given ID.
	Get(ctx context.Context, id string) (*Board, error)

	// Create a new board. When creating a new board ID may not be set.
	Create(ctx context.Context, b *Board) (*Board, error)

	// Update an existing board.
	Update(ctx context.Context, b *Board) (*Board, error)

	// Delete a board.
	Delete(ctx context.Context, id string) error
}

// boards implements Boards.
type boards struct {
	client *Client
}

// Compile-time proof of interface implementation by type boards.
var _ Boards = (*boards)(nil)

// Board represents a Honeycomb board.
//
// API docs: https://docs.honeycomb.io/api/boards-api/#fields-on-a-board
type Board struct {
	ID string `json:"id,omitempty"`

	// Name of the board, this is displayed in the Honeycomb UI. This field is
	// required.
	Name string `json:"name"`
	// Description of the board.
	Description string `json:"description,omitempty"`
	// How the board should be displayed in the UI, defaults to "list".
	Style BoardStyle `json:"style,omitempty"`
	// A list of queries displayed on the board, in order of appearance.
	Queries []BoardQuery `json:"queries"`
}

// BoardStyle determines how a Board should be displayed within the Honeycomb UI.
type BoardStyle string

// Declaration of board styles.
const (
	BoardStyleList   BoardStyle = "list"
	BoardStyleVisual BoardStyle = "visual"
)

// BoardStyles returns an exhaustive list of board styles.
func BoardStyles() []BoardStyle {
	return []BoardStyle{BoardStyleList, BoardStyleVisual}
}

// BoardQuery represents a query that is part of a board.
type BoardQuery struct {
	Caption string `json:"caption,omitempty"`
	// This field is required.
	Dataset string `json:"dataset"`
	// This field is required.
	Query QuerySpec `json:"query"`
}

func (s *boards) List(ctx context.Context) ([]Board, error) {
	var b []Board
	err := s.client.performRequest(ctx, "GET", "/1/boards", nil, &b)
	return b, err
}

func (s *boards) Get(ctx context.Context, ID string) (*Board, error) {
	var b Board
	err := s.client.performRequest(ctx, "GET", fmt.Sprintf("/1/boards/%s", ID), nil, &b)
	return &b, err
}

func (s *boards) Create(ctx context.Context, data *Board) (*Board, error) {
	var b Board
	err := s.client.performRequest(ctx, "POST", "/1/boards", data, &b)
	return &b, err
}

func (s *boards) Update(ctx context.Context, data *Board) (*Board, error) {
	var b Board
	err := s.client.performRequest(ctx, "PUT", fmt.Sprintf("/1/boards/%s", data.ID), data, &b)
	return &b, err
}

func (s *boards) Delete(ctx context.Context, id string) error {
	return s.client.performRequest(ctx, "DELETE", fmt.Sprintf("/1/boards/%s", id), nil, nil)
}
