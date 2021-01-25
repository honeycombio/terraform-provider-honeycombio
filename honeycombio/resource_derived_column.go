package honeycombio

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/kvrhdn/go-honeycombio"
)

func newDerivedColumn() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDerivedColumnCreate,
		ReadContext:   resourceDerivedColumnRead,
		UpdateContext: resourceDerivedColumnUpdate,
		DeleteContext: resourceDerivedColumnDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceDerivedColumnImport,
		},

		Schema: map[string]*schema.Schema{
			"alias": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"expression": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"dataset": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceDerivedColumnImport(ctx context.Context, d *schema.ResourceData, i interface{}) ([]*schema.ResourceData, error) {
	// import ID is of the format <dataset>/<derived column alias>
	// note that the dataset name can also contain '/'
	idSegments := strings.Split(d.Id(), "/")
	if len(idSegments) < 2 {
		return nil, fmt.Errorf("invalid import ID, supplied ID must be written as <dataset>/<derived column alias>")
	}

	dataset := strings.Join(idSegments[0:len(idSegments)-1], "/")
	alias := idSegments[len(idSegments)-1]

	d.Set("alias", alias)
	d.Set("dataset", dataset)
	return []*schema.ResourceData{d}, nil
}

func resourceDerivedColumnCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	dataset := d.Get("dataset").(string)
	derivedColumn := readDerivedColumn(d)

	derivedColumn, err := client.DerivedColumns.Create(ctx, dataset, derivedColumn)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("alias", derivedColumn.Alias)
	return resourceDerivedColumnRead(ctx, d, meta)
}

func resourceDerivedColumnRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	dataset := d.Get("dataset").(string)

	derivedColumn, err := client.DerivedColumns.GetByAlias(ctx, dataset, d.Get("alias").(string))
	if err != nil {
		if err == honeycombio.ErrNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId(derivedColumn.ID)
	d.Set("alias", derivedColumn.Alias)
	d.Set("expression", derivedColumn.Expression)
	d.Set("description", derivedColumn.Description)
	return nil
}

func resourceDerivedColumnUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	dataset := d.Get("dataset").(string)
	derivedColumn := readDerivedColumn(d)

	derivedColumn, err := client.DerivedColumns.Update(ctx, dataset, derivedColumn)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("alias", derivedColumn.Alias)
	return resourceDerivedColumnRead(ctx, d, meta)
}

func resourceDerivedColumnDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	dataset := d.Get("dataset").(string)

	err := client.DerivedColumns.Delete(ctx, dataset, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func readDerivedColumn(d *schema.ResourceData) *honeycombio.DerivedColumn {
	return &honeycombio.DerivedColumn{
		Alias:       d.Get("alias").(string),
		Expression:  d.Get("expression").(string),
		Description: d.Get("description").(string),
	}
}
