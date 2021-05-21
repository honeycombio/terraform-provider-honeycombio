package honeycombio

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/kvrhdn/go-honeycombio"
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
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"style": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "list",
				ValidateFunc: validation.StringInSlice(boardStyleStrings(), false),
			},
			"query": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"caption": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"query_style": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringInSlice(boardQueryStyleStrings(), false),
						},
						"dataset": {
							Type:     schema.TypeString,
							Required: true,
						},
						"query_json": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: validateQueryJSON(),
						},
						"query_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"query_annotation_id": {
							Type:     schema.TypeString,
							Optional: true,
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
	if err != nil {
		if err == honeycombio.ErrNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	// API returns "" for filterCombination if set to the default value "AND"
	// To keep the Terraform config simple, we'll explicitly set "AND" ourself
	for i := range b.Queries {
		q := &b.Queries[i]
		if q.Query.FilterCombination == "" {
			q.Query.FilterCombination = honeycombio.FilterCombinationAnd
		}
	}

	d.SetId(b.ID)
	d.Set("name", b.Name)
	d.Set("description", b.Description)
	d.Set("style", b.Style)

	queries := make([]map[string]interface{}, len(b.Queries))

	for i, q := range b.Queries {
		queryJSON, err := encodeQuery(q.Query)
		if err != nil {
			return diag.FromErr(err)
		}

		queries[i] = map[string]interface{}{
			"caption":             q.Caption,
			"query_style":         q.QueryStyle,
			"dataset":             q.Dataset,
			"query_json":          queryJSON,
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
	var queries []honeycombio.BoardQuery

	qs := d.Get("query").([]interface{})
	for _, q := range qs {
		m := q.(map[string]interface{})

		var query *honeycombio.QuerySpec
		if m["query_json"] != nil {
			query = new(honeycombio.QuerySpec)
			err := json.Unmarshal([]byte(m["query_json"].(string)), query)
			if err != nil {
				return nil, err
			}
		}

		queries = append(queries, honeycombio.BoardQuery{
			Caption:           m["caption"].(string),
			QueryStyle:        honeycombio.BoardQueryStyle(m["query_style"].(string)),
			Dataset:           m["dataset"].(string),
			Query:             query,
			QueryID:           m["query_id"].(string),
			QueryAnnotationID: m["query_annotation_id"].(string),
		})
	}

	board := &honeycombio.Board{
		ID:          d.Id(),
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Style:       honeycombio.BoardStyle(d.Get("style").(string)),
		Queries:     queries,
	}
	return board, nil
}
