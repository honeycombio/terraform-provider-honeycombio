package honeycombio

import (
	"context"
	"fmt"
)

// Compile-time proof of interface implementation.
var _ Boards = (*boards)(nil)

// Boards describes (some of) the board related methods that Honeycomb supports.
type Boards interface {
	// List all boards.
	List(ctx context.Context) ([]Board, error)

	// Get a board by its ID. Returns nil, ErrNotFound if there is no board
	// with the given ID.
	Get(ctx context.Context, id string) (*Board, error)

	// Create a new board.
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

// Board represents a Honeycomb board, as described by https://docs.honeycomb.io/api/boards-api/#fields-on-a-board
type Board struct {
	// The generated ID of the board.  This should not be specified by the user in the creation request.
	ID string `json:"id,omitempty"`
	// The (required) board's name displayed in the UI
	Name string `json:"name"`
	// The description of the board
	Description string `json:"description,omitempty"`
	// How the board should be displayed in the UI, either "list" (the default) or "visual"
	Style BoardStyle `json:"style,omitempty"`
	// A list of queries to display on the board in order of appearance
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

// BoardQuery represents are query that is part of a board.
type BoardQuery struct {
	Caption string    `json:"caption,omitempty"`
	Dataset string    `json:"dataset"`
	Query   QuerySpec `json:"query"`
}

func (s *boards) List(ctx context.Context) ([]Board, error) {
	req, err := s.client.newRequest(ctx, "GET", "/1/boards", nil)
	if err != nil {
		return nil, err
	}

	var b []Board
	err = s.client.do(req, &b)
	return b, err
}

func (s *boards) Get(ctx context.Context, ID string) (*Board, error) {
	req, err := s.client.newRequest(ctx, "GET", "/1/boards/"+ID, nil)
	if err != nil {
		return nil, err
	}

	var b Board
	err = s.client.do(req, &b)
	return &b, err
}

func (s *boards) Create(ctx context.Context, data *Board) (*Board, error) {
	req, err := s.client.newRequest(ctx, "POST", "/1/boards", data)
	if err != nil {
		return nil, err
	}

	var b Board
	err = s.client.do(req, &b)
	return &b, err
}

func (s *boards) Update(ctx context.Context, data *Board) (*Board, error) {
	req, err := s.client.newRequest(ctx, "PUT", fmt.Sprintf("/1/boards/%s", data.ID), data)
	if err != nil {
		return nil, err
	}

	var b Board
	err = s.client.do(req, &b)
	return &b, err
}

func (s *boards) Delete(ctx context.Context, id string) error {
	req, err := s.client.newRequest(ctx, "DELETE", fmt.Sprintf("/1/boards/%s", id), nil)
	if err != nil {
		return nil
	}

	return s.client.do(req, nil)
}
