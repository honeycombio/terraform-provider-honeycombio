package honeycombio

import "fmt"

// Compile-time proof of interface implementation.
var _ Boards = (*boards)(nil)

// Boards describes (some of) the board related methods that Honeycomb supports.
type Boards interface {
	// List all boards.
	List() ([]Board, error)

	// Get a board by its ID. Returns nil, ErrNotFound if there is no board
	// with the given ID.
	Get(id string) (*Board, error)

	// Create a new board.
	Create(b *Board) (*Board, error)

	// Update an existing board.
	Update(b *Board) (*Board, error)

	// Delete a board.
	Delete(id string) error
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

// List of available board styles.
const (
	BoardStyleList   BoardStyle = "list"
	BoardStyleVisual BoardStyle = "visual"
)

// BoardQuery represents are query that is part of a board.
type BoardQuery struct {
	Caption string    `json:"caption,omitempty"`
	Dataset string    `json:"dataset"`
	Query   QuerySpec `json:"query"`
}

func (s *boards) List() ([]Board, error) {
	req, err := s.client.newRequest("GET", "/1/boards", nil)
	if err != nil {
		return nil, err
	}

	var b []Board
	err = s.client.do(req, &b)
	return b, err
}

func (s *boards) Get(ID string) (*Board, error) {
	req, err := s.client.newRequest("GET", "/1/boards/"+ID, nil)
	if err != nil {
		return nil, err
	}

	var b Board
	err = s.client.do(req, &b)
	return &b, err
}

func (s *boards) Create(data *Board) (*Board, error) {
	req, err := s.client.newRequest("POST", "/1/boards", data)
	if err != nil {
		return nil, err
	}

	var b Board
	err = s.client.do(req, &b)
	return &b, err
}

func (s *boards) Update(data *Board) (*Board, error) {
	req, err := s.client.newRequest("PUT", fmt.Sprintf("/1/boards/%s", data.ID), data)
	if err != nil {
		return nil, err
	}

	var b Board
	err = s.client.do(req, &b)
	return &b, err
}

func (s *boards) Delete(id string) error {
	req, err := s.client.newRequest("DELETE", fmt.Sprintf("/1/boards/%s", id), nil)
	if err != nil {
		return nil
	}

	return s.client.do(req, nil)
}
