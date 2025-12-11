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

	// Get a board by its ID.
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

	// Type of the board, this controls the display between flexible and classic boards in the Honeycomb UI.
	// Defaults to "classic".
	BoardType BoardType `json:"type,omitempty"`

	// Layout generation controls how the board layout is generated.
	// Defaults to "manual".
	LayoutGeneration LayoutGeneration `json:"layout_generation,omitempty"`

	// Board panels are the individual panels that make up a board. Each panel can
	// be a query or an SLO panel. The panels are laid out in a grid.
	Panels []BoardPanel `json:"panels,omitempty"`

	// Name of the board, this is displayed in the Honeycomb UI. This field is
	// required.
	Name string `json:"name"`
	// Description of the board.
	Description string `json:"description,omitempty"`
	// Links returned by the board API for the Board
	Links BoardLinks `json:"links,omitempty"`
	// A list of tags to organize the Board, for flexible boards only
	Tags []Tag `json:"tags"`
	// A list of preset filters for the board.
	PresetFilters *[]PresetFilter `json:"preset_filters,omitempty"`
}

// PresetFilter represents a single combination of column name and alias for a preset filter on a board.
type PresetFilter struct {
	Column string `json:"column,omitempty"`
	Alias  string `json:"alias,omitempty"`
}

// BoardPanel represents a single panel on a board.
type BoardPanel struct {
	PanelType     BoardPanelType     `json:"type,omitempty"` // "query" or "slo"
	PanelPosition BoardPanelPosition `json:"position,omitempty"`

	QueryPanel *BoardQueryPanel `json:"query_panel,omitempty"`
	SLOPanel   *BoardSLOPanel   `json:"slo_panel,omitempty"`
	TextPanel  *BoardTextPanel  `json:"text_panel,omitempty"`
}

func (b *BoardPanel) IsBlank() bool {
	return b.PanelPosition.X == 0 && b.PanelPosition.Y == 0 && b.PanelPosition.Height == 0 && b.PanelPosition.Width == 0
}

type BoardPanelType string

const (
	BoardPanelTypeQuery BoardPanelType = "query"
	BoardPanelTypeSLO   BoardPanelType = "slo"
	BoardPanelTypeText  BoardPanelType = "text"
)

type BoardPanelPosition struct {
	X      int `json:"x_coordinate"`
	Y      int `json:"y_coordinate"`
	Height int `json:"height"`
	Width  int `json:"width"`
}

type BoardQueryPanel struct {
	Dataset               string                           `json:"dataset,omitempty"`
	QueryID               string                           `json:"query_id,omitempty"`
	QueryAnnotationID     string                           `json:"query_annotation_id,omitempty"`
	VisualizationSettings *BoardQueryVisualizationSettings `json:"visualization_settings,omitempty"`
	Style                 BoardQueryStyle                  `json:"query_style,omitempty"`
}

type BoardQueryVisualizationSettings struct {
	UseUTCXAxis          bool             `json:"utc_xaxis,omitempty"`
	HideMarkers          bool             `json:"hide_markers,omitempty"`
	HideHovers           bool             `json:"hide_hovers,omitempty"`
	PreferOverlaidCharts bool             `json:"overlaid_charts,omitempty"`
	HideCompare          bool             `json:"hide_compare,omitempty"`
	Charts               []*ChartSettings `json:"charts,omitempty"`
}

type ChartSettings struct {
	ChartType         string `json:"chart_type,omitempty"` // "default", "line", "stacked", "stat", "tsbar"
	ChartIndex        int    `json:"chart_index"`
	OmitMissingValues bool   `json:"omit_missing_values,omitempty"`
	UseLogScale       bool   `json:"log_scale,omitempty"`
}

type BoardSLOPanel struct {
	SLOID string `json:"slo_id,omitempty"`
}

type BoardTextPanel struct {
	Content string `json:"content,omitempty"`
}

type BoardType string

const (
	BoardTypeFlexible BoardType = "flexible"
)

type LayoutGeneration string

const (
	LayoutGenerationManual LayoutGeneration = "manual"
	LayoutGenerationAuto   LayoutGeneration = "auto"
)

// BoardLinks represents links returned by the board API.
type BoardLinks struct {
	// URL For accessing the board
	BoardURL string `json:"board_url,omitempty"`
}

// BoardQueryStyle determines how a query should be displayed on the board.
type BoardQueryStyle string

// Declaration of board query styles.
const (
	BoardQueryStyleGraph BoardQueryStyle = "graph"
	BoardQueryStyleTable BoardQueryStyle = "table"
	BoardQueryStyleCombo BoardQueryStyle = "combo"
)

func (s *boards) List(ctx context.Context) ([]Board, error) {
	var b []Board
	err := s.client.Do(ctx, "GET", "/1/boards", nil, &b)
	return b, err
}

func (s *boards) Get(ctx context.Context, ID string) (*Board, error) {
	var b Board
	err := s.client.Do(ctx, "GET", fmt.Sprintf("/1/boards/%s", ID), nil, &b)
	return &b, err
}

func (s *boards) Create(ctx context.Context, data *Board) (*Board, error) {
	var b Board
	err := s.client.Do(ctx, "POST", "/1/boards", data, &b)
	return &b, err
}

func (s *boards) Update(ctx context.Context, data *Board) (*Board, error) {
	var b Board
	err := s.client.Do(ctx, "PUT", fmt.Sprintf("/1/boards/%s", data.ID), data, &b)
	return &b, err
}

func (s *boards) Delete(ctx context.Context, id string) error {
	return s.client.Do(ctx, "DELETE", fmt.Sprintf("/1/boards/%s", id), nil, nil)
}
