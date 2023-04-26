package honeycombio

import (
	"context"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
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
	client := meta.(*honeycombio.Client)

	dataset := d.Get("dataset").(string)

	columns, err := client.Columns.List(ctx, dataset)
	if err != nil {
		return diag.FromErr(err)
	}

	value, ok := d.GetOk("starts_with")
	if ok {
		startsWith := value.(string)

		for i := len(columns) - 1; i >= 0; i-- {
			if !strings.HasPrefix(columns[i].KeyName, startsWith) {
				columns = append(columns[:i], columns[i+1:]...)
			}
		}
	}

	names := make([]string, len(columns))
	for i, d := range columns {
		names[i] = d.KeyName
	}
	d.Set("names", names)

	d.SetId(strconv.Itoa(hashcode.String(strings.Join(names, ","))))
	return nil
}
