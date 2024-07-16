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

	dataset := d.Get("dataset").(string)
	matchName := d.Get("name").(string)

	column, err := client.Columns.GetByKeyName(ctx, dataset, matchName)
	if err != nil {
		return diagFromErr(err)
	}

	d.SetId(column.ID)
	d.Set("name", column.KeyName)
	d.Set("hidden", column.Hidden)
	d.Set("type", column.Type)
	d.Set("description", column.Description)
	d.Set("last_written_at", column.LastWrittenAt.UTC().Format(time.RFC3339))
	d.Set("created_at", column.CreatedAt.UTC().Format(time.RFC3339))
	d.Set("updated_at", column.UpdatedAt.UTC().Format(time.RFC3339))

	return nil
}
