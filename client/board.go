package client

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
	// The number of columns to be layed out when displaying the board.
	// Defaults to "multi".
	//
	// n.b. 'list' style boards cannot specify a column layout
	ColumnLayout BoardColumnStyle `json:"column_layout,omitempty"`
	// How the board should be displayed in the UI, defaults to "list".
	Style BoardStyle `json:"style,omitempty"`
	// Links returned by the board API for the Board
	Links BoardLinks `json:"links,omitempty"`
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

type BoardColumnStyle string

const (
	BoardColumnStyleMulti  BoardColumnStyle = "multi"
	BoardColumnStyleSingle BoardColumnStyle = "single"
)

// BoardLinks represents links returned by the board API.
type BoardLinks struct {
	// URL For accessing the board
	BoardURL string `json:"board_url,omitempty"`
}

// BoardQuery represents a query that is part of a board.
type BoardQuery struct {
	Caption string `json:"caption,omitempty"`
	// Defaults to graph.
	QueryStyle BoardQueryStyle `json:"query_style,omitempty"`
	// Dataset is no longer required
	Dataset string `json:"dataset,omitempty"`
	// QueryID is required
	QueryID string `json:"query_id,omitempty"`
	// Optional
	QueryAnnotationID string `json:"query_annotation_id,omitempty"`
	// Optional
	GraphSettings BoardGraphSettings `json:"graph_settings"`
}

// BoardQueryStyle determines how a query should be displayed on the board.
type BoardQueryStyle string

// Declaration of board query styles.
const (
	BoardQueryStyleGraph BoardQueryStyle = "graph"
	BoardQueryStyleTable BoardQueryStyle = "table"
	BoardQueryStyleCombo BoardQueryStyle = "combo"
)

// BoardGraphSettings represents the display settings for an individual graph in a board.
type BoardGraphSettings struct {
	OmitMissingValues    bool `json:"omit_missing_values,omitempty"`
	UseStackedGraphs     bool `json:"stacked_graphs,omitempty"`
	UseLogScale          bool `json:"log_scale,omitempty"`
	UseUTCXAxis          bool `json:"utc_xaxis,omitempty"`
	HideMarkers          bool `json:"hide_markers,omitempty"`
	PreferOverlaidCharts bool `json:"overlaid_charts,omitempty"`
}

// BoardQueryStyles returns an exhaustive list of board query styles.
func BoardQueryStyles() []BoardQueryStyle {
	return []BoardQueryStyle{BoardQueryStyleGraph, BoardQueryStyleTable, BoardQueryStyleCombo}
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
