package honeycombio

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
)

func newColumn() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceColumnCreate,
		ReadContext:   resourceColumnRead,
		UpdateContext: resourceColumnUpdate,
		DeleteContext: schema.NoopContext,
		Importer: &schema.ResourceImporter{
			StateContext: resourceColumnImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     false, // will be true when key_name is removed
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
				AtLeastOneOf: []string{"key_name", "name"},
			},
			"key_name": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				Computed:      true,
				Deprecated:    "Please set `name` instead.",
				ConflictsWith: []string{"name"},
				ValidateFunc:  validation.StringLenBetween(1, 255),
			},
			"hidden": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "string",
				ValidateFunc: validation.StringInSlice(columnTypeStrings(), false),
			},
			"dataset": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
				Required: false,
				Optional: false,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
				Required: false,
				Optional: false,
			},
			"last_written_at": {
				Type:     schema.TypeString,
				Computed: true,
				Required: false,
				Optional: false,
			},
		},
	}
}

func resourceColumnImport(ctx context.Context, d *schema.ResourceData, i interface{}) ([]*schema.ResourceData, error) {
	// import ID is of the format <dataset>/<column name>
	// note that the dataset name can also contain '/'
	idSegments := strings.Split(d.Id(), "/")
	if len(idSegments) < 2 {
		return nil, fmt.Errorf("invalid import ID, supplied ID must be written as <dataset>/<column name>")
	}

	dataset := strings.Join(idSegments[0:len(idSegments)-1], "/")
	name := idSegments[len(idSegments)-1]

	d.Set("name", name)
	d.Set("dataset", dataset)
	return []*schema.ResourceData{d}, nil
}

func resourceColumnCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	dataset := d.Get("dataset").(string)

	column, err := client.Columns.Create(ctx, dataset, expandColumn(d))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(column.ID)
	return resourceColumnRead(ctx, d, meta)
}

func resourceColumnRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	dataset := d.Get("dataset").(string)
	// if name is not set, try to get key_name.
	// The schema requires one or the other to be set
	columnName := d.Get("name").(string)
	if columnName == "" {
		columnName = d.Get("key_name").(string)
	}

	// we read by name here to faciliate importing by name instead of ID
	column, err := client.Columns.GetByKeyName(ctx, dataset, columnName)
	if err == honeycombio.ErrNotFound {
		d.SetId("")
		return nil
	} else if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(column.ID)
	d.Set("name", column.KeyName)
	d.Set("key_name", column.KeyName)
	d.Set("hidden", column.Hidden)
	d.Set("description", column.Description)
	d.Set("type", column.Type)
	d.Set("created_at", column.CreatedAt.UTC().Format(time.RFC3339))
	d.Set("updated_at", column.CreatedAt.UTC().Format(time.RFC3339))
	d.Set("last_written_at", column.CreatedAt.UTC().Format(time.RFC3339))
	return nil
}

func resourceColumnUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	dataset := d.Get("dataset").(string)

	column, err := client.Columns.Update(ctx, dataset, expandColumn(d))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(column.ID)
	return resourceColumnRead(ctx, d, meta)
}

func expandColumn(d *schema.ResourceData) *honeycombio.Column {
	// if name is not set, try to get key_name.
	// The schema requires one or the other to be set
	columnName := d.Get("name").(string)
	if columnName == "" {
		columnName = d.Get("key_name").(string)
	}

	return &honeycombio.Column{
		ID:          d.Id(),
		KeyName:     columnName,
		Hidden:      honeycombio.ToPtr(d.Get("hidden").(bool)),
		Description: d.Get("description").(string),
		Type:        honeycombio.ToPtr(honeycombio.ColumnType(d.Get("type").(string))),
	}
}
