package honeycombio

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceHoneycombioColumn() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHoneycombioColumnRead,

		Description: `
'honeycombio_column' data source provides details about a specific column in a dataset, matched by name.
`,

		Schema: map[string]*schema.Schema{
			"dataset": {
				Type:        schema.TypeString,
				Optional:    false,
				Required:    true,
				Description: "The dataset this column is associated with",
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    false,
				Required:    true,
				Description: "Name of the column",
			},
			// Generated attributes
			"hidden": {
				Type:        schema.TypeBool,
				Required:    false,
				Optional:    false,
				Computed:    true,
				Description: "Whether the column is hidden",
			},
			"type": {
				Type:        schema.TypeString,
				Required:    false,
				Optional:    false,
				Computed:    true,
				Description: "The type of column, allowed types are `string`, `integer`, `float`, and `boolean`.",
			},
			"description": {
				Type:        schema.TypeString,
				Required:    false,
				Optional:    false,
				Computed:    true,
				Description: "A description of the column",
			},
			"last_written_at": {
				Type:        schema.TypeString,
				Required:    false,
				Optional:    false,
				Computed:    true,
				Description: "The last time the column was written to",
			},
			"created_at": {
				Type:        schema.TypeString,
				Required:    false,
				Optional:    false,
				Computed:    true,
				Description: "The time the column was created",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Required:    false,
				Optional:    false,
				Computed:    true,
				Description: "The time the column was last updated",
			},
		},
	}
}

func dataSourceHoneycombioColumnRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := getConfiguredClient(meta)
	if err != nil {
		return diagFromErr(err)
	}

	dataset, ok := d.Get("dataset").(string)
	if !ok {
		return diag.Errorf("dataset must be a string")
	}
	matchName, ok := d.Get("name").(string)
	if !ok {
		return diag.Errorf("name must be a string")
	}

	column, err := client.Columns.GetByKeyName(ctx, dataset, matchName)
	if err != nil {
		return diagFromErr(err)
	}

	d.SetId(column.ID)
	_ = d.Set("name", column.KeyName)
	_ = d.Set("hidden", column.Hidden)
	_ = d.Set("type", column.Type)
	_ = d.Set("description", column.Description)
	_ = d.Set("last_written_at", column.LastWrittenAt.UTC().Format(time.RFC3339))
	_ = d.Set("created_at", column.CreatedAt.UTC().Format(time.RFC3339))
	_ = d.Set("updated_at", column.UpdatedAt.UTC().Format(time.RFC3339))

	return nil
}
