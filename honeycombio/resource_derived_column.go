package honeycombio

import (
	"context"
	"errors"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
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
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"expression": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 4095),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
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
	importID := d.Id()
	idx := strings.Index(importID, "/")
	if idx < 0 {
		return nil, errors.New("invalid import ID, supplied ID must be written as <dataset>/<derived column alias>")
	}
	dataset := importID[:idx]
	alias := importID[idx+1:]

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
		return diagFromErr(err)
	}

	d.Set("alias", derivedColumn.Alias)
	return resourceDerivedColumnRead(ctx, d, meta)
}

func resourceDerivedColumnRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	dataset := d.Get("dataset").(string)

	var detailedErr honeycombio.DetailedError
	derivedColumn, err := client.DerivedColumns.GetByAlias(ctx, dataset, d.Get("alias").(string))
	if errors.As(err, &detailedErr) {
		if detailedErr.IsNotFound() {
			d.SetId("")
			return nil
		} else {
			return diagFromDetailedErr(detailedErr)
		}
	} else if err != nil {
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
		return diagFromErr(err)
	}

	d.Set("alias", derivedColumn.Alias)
	return resourceDerivedColumnRead(ctx, d, meta)
}

func resourceDerivedColumnDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	dataset := d.Get("dataset").(string)

	err := client.DerivedColumns.Delete(ctx, dataset, d.Id())
	if err != nil {
		return diagFromErr(err)
	}
	return nil
}

func readDerivedColumn(d *schema.ResourceData) *honeycombio.DerivedColumn {
	return &honeycombio.DerivedColumn{
		ID:          d.Id(),
		Alias:       d.Get("alias").(string),
		Expression:  d.Get("expression").(string),
		Description: d.Get("description").(string),
	}
}
