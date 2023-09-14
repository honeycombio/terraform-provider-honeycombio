package honeycombio

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
)

func newBoard() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceBoardCreate,
		ReadContext:   resourceBoardRead,
		UpdateContext: resourceBoardUpdate,
		DeleteContext: resourceBoardDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The name of the Board.",
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "A description for the Board. Supports Markdown.",
				ValidateFunc: validation.StringLenBetween(1, 1023),
			},
			"column_layout": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				Description:  "The number of columns to layout on the Board.",
				ValidateFunc: validation.StringInSlice([]string{"multi", "single"}, false),
			},
			"style": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "visual",
				Description:  "How the Board should be displayed in the UI.",
				Deprecated:   "All Boards now are displayed visually in the UI. Setting this value will have no effect.",
				ValidateFunc: validation.StringInSlice(boardStyleStrings(), false),
			},
			"board_url": {
				Type:        schema.TypeString,
				Description: "The URL to the Board in the Honeycomb UI.",
				Required:    false,
				Optional:    false,
				Computed:    true,
			},
			"query": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "A Query to be added to the Board.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"caption": {
							Type:         schema.TypeString,
							Optional:     true,
							Description:  "Descriptive text to contextualize the Query within the Board. Supports Markdown.",
							ValidateFunc: validation.StringLenBetween(1, 1023),
						},
						"query_style": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							Description:  "How the Query should be displayed within the Board.",
							ValidateFunc: validation.StringInSlice(boardQueryStyleStrings(), false),
						},
						"graph_settings": {
							Type:     schema.TypeList,
							MinItems: 1,
							MaxItems: 1,
							Computed: true,
							Optional: true,
							Description: `Manages the settings for this query's graph on the board.
See [Graph Settings](https://docs.honeycomb.io/working-with-your-data/graph-settings/) in the documentation for more information.`,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"log_scale": {
										Type:        schema.TypeBool,
										Computed:    true,
										Optional:    true,
										Description: "Set the graph's Y axis to Log scale.",
									},
									"omit_missing_values": {
										Type:        schema.TypeBool,
										Computed:    true,
										Optional:    true,
										Description: "Enable interpolatation between datapoints when the interveneing time buckets have no matching events.",
									},
									"hide_markers": {
										Type:        schema.TypeBool,
										Computed:    true,
										Optional:    true,
										Description: "Disable the overlay of Markers on the graph.",
									},
									"stacked_graphs": {
										Type:        schema.TypeBool,
										Computed:    true,
										Optional:    true,
										Description: "Enable the display of groups as stacked colored area under their line graphs.",
									},
									"utc_xaxis": {
										Type:        schema.TypeBool,
										Computed:    true,
										Optional:    true,
										Description: "Set the graph's X axis to UTC.",
									},
									"overlaid_charts": {
										Type:        schema.TypeBool,
										Computed:    true,
										Optional:    true,
										Description: "Allow charts to be overlaid when using supported Visualize operators.",
									},
								},
							},
						},
						"dataset": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "The Dataset this Query is associated with.",
							Deprecated:  "Board Queries no longer require the dataset as they rely on the provided Query ID's dataset.",
						},
						"query_id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The ID of the Query to run.",
						},
						"query_annotation_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The ID of the Query Annotation to associate with this Query.",
						},
					},
				},
			},
		},
	}
}

func resourceBoardCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	b, err := expandBoard(d)
	if err != nil {
		return diag.FromErr(err)
	}

	b, err = client.Boards.Create(ctx, b)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(b.ID)
	return resourceBoardRead(ctx, d, meta)
}

func resourceBoardRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	b, err := client.Boards.Get(ctx, d.Id())
	if err == honeycombio.ErrNotFound {
		d.SetId("")
		return nil
	} else if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(b.ID)
	d.Set("name", b.Name)
	d.Set("description", b.Description)
	d.Set("style", b.Style)
	d.Set("column_layout", b.ColumnLayout)
	d.Set("board_url", b.Links.BoardURL)

	queries := make([]map[string]interface{}, len(b.Queries))

	for i, q := range b.Queries {
		queries[i] = map[string]interface{}{
			"caption":             q.Caption,
			"query_style":         q.QueryStyle,
			"dataset":             q.Dataset,
			"graph_settings":      flattenBoardQueryGraphSettings(q.GraphSettings),
			"query_id":            q.QueryID,
			"query_annotation_id": q.QueryAnnotationID,
		}
	}

	d.Set("query", queries)

	return nil
}

func resourceBoardUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	b, err := expandBoard(d)
	if err != nil {
		return diag.FromErr(err)
	}

	b, err = client.Boards.Update(ctx, b)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(b.ID)
	return resourceBoardRead(ctx, d, meta)
}

func resourceBoardDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	err := client.Boards.Delete(ctx, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func expandBoard(d *schema.ResourceData) (*honeycombio.Board, error) {
	var err error
	var queries []honeycombio.BoardQuery

	qs := d.Get("query").([]interface{})
	for _, q := range qs {
		m := q.(map[string]interface{})

		var graphSettings honeycombio.BoardGraphSettings
		if v, ok := m["graph_settings"]; ok {
			graphSettings, err = expandBoardQueryGraphSettings(v)
			if err != nil {
				return nil, err
			}
		}
		queries = append(queries, honeycombio.BoardQuery{
			Caption:           m["caption"].(string),
			QueryStyle:        honeycombio.BoardQueryStyle(m["query_style"].(string)),
			GraphSettings:     graphSettings,
			Dataset:           m["dataset"].(string),
			QueryID:           m["query_id"].(string),
			QueryAnnotationID: m["query_annotation_id"].(string),
		})
	}

	board := &honeycombio.Board{
		ID:           d.Id(),
		Name:         d.Get("name").(string),
		Description:  d.Get("description").(string),
		Style:        honeycombio.BoardStyle(d.Get("style").(string)),
		ColumnLayout: honeycombio.BoardColumnStyle(d.Get("column_layout").(string)),
		Queries:      queries,
	}

	if board.Style == honeycombio.BoardStyleList && board.ColumnLayout != "" {
		return nil, errors.New("list style boards cannot specify a column layout")
	}
	return board, nil
}

func expandBoardQueryGraphSettings(gs interface{}) (honeycombio.BoardGraphSettings, error) {
	graphSettings := honeycombio.BoardGraphSettings{}
	raw := gs.([]interface{})
	if len(raw) == 0 {
		return graphSettings, nil
	}
	if len(raw) > 1 {
		return graphSettings, errors.New("got more than one set of graph settings?")
	}
	s, ok := raw[0].(map[string]interface{})
	if !ok {
		return graphSettings, nil
	}

	if v, ok := s["log_scale"].(bool); ok && v {
		graphSettings.UseLogScale = true
	}
	if v, ok := s["omit_missing_values"].(bool); ok && v {
		graphSettings.OmitMissingValues = true
	}
	if v, ok := s["hide_markers"].(bool); ok && v {
		graphSettings.HideMarkers = true
	}
	if v, ok := s["stacked_graphs"].(bool); ok && v {
		graphSettings.UseStackedGraphs = true
	}
	if v, ok := s["utc_xaxis"].(bool); ok && v {
		graphSettings.UseUTCXAxis = true
	}
	if v, ok := s["overlaid_charts"].(bool); ok && v {
		graphSettings.PreferOverlaidCharts = true
	}

	return graphSettings, nil
}

func flattenBoardQueryGraphSettings(gs honeycombio.BoardGraphSettings) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)

	result = append(result, map[string]interface{}{
		"hide_markers":        gs.HideMarkers,
		"log_scale":           gs.UseLogScale,
		"omit_missing_values": gs.OmitMissingValues,
		"stacked_graphs":      gs.UseStackedGraphs,
		"utc_xaxis":           gs.UseUTCXAxis,
		"overlaid_charts":     gs.PreferOverlaidCharts,
	})

	return result
}
