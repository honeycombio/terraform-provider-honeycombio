package honeycombio

import (
	"context"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/hashcode"
)

func dataSourceHoneycombioColumns() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHoneycombioColumnsRead,

		Schema: map[string]*schema.Schema{
			"dataset": {
				Type:     schema.TypeString,
				Optional: false,
				Required: true,
			},
			"starts_with": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"names": {
				Type:     schema.TypeList,
				Computed: true,
				Optional: false,
				Required: false,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dataSourceHoneycombioColumnsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := getConfiguredClient(meta)
	if err != nil {
		return diagFromErr(err)
	}

	dataset, ok := d.Get("dataset").(string)
	if !ok {
		return diag.Errorf("dataset must be a string")
	}

	columns, err := client.Columns.List(ctx, dataset)
	if err != nil {
		return diagFromErr(err)
	}

	var startsWith string
	if value, ok := d.GetOk("starts_with"); ok {
		startsWith, ok = value.(string)
		if !ok {
			return diag.Errorf("starts_with must be a string")
		}
	}

	names := make([]string, 0, len(columns))
	for _, column := range columns {
		if startsWith != "" && !strings.HasPrefix(column.KeyName, startsWith) {
			continue
		}
		names = append(names, column.KeyName)
	}
	_ = d.Set("names", names)

	d.SetId(strconv.Itoa(hashcode.String(strings.Join(names, ","))))
	return nil
}
