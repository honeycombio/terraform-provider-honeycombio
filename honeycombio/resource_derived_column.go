package honeycombio

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	dcparser "github.com/honeycombio/honeycomb-derived-column-validator/pkg/parser"

	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/honeycombio/internal/verify"
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
				Description:  "The alias of the derived column. Must be unique within the dataset or environment.",
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"expression": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The formula of the derived column. See [Derived Column Syntax](https://docs.honeycomb.io/working-with-your-data/customizing-your-query/derived-columns/#derived-column-syntax).",
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 4095),
					func(i interface{}, k string) ([]string, []error) {
						v, ok := i.(string)
						if !ok {
							return nil, []error{fmt.Errorf("expected type of %s to be string", k)}
						}

						if _, err := dcparser.ANTLRParse(v, false); err != nil {
							return nil, []error{fmt.Errorf("invalid derived column syntax: %s", err)}
						}

						return nil, nil
					},
				),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "A description of the derived column.",
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"dataset": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Description:      "The dataset this derived column belongs to. If not set, it will be Environment-wide.",
				DiffSuppressFunc: verify.SuppressEquivEnvWideDataset,
			},
		},
	}
}

func resourceDerivedColumnImport(ctx context.Context, d *schema.ResourceData, i interface{}) ([]*schema.ResourceData, error) {
	dataset, alias, found := strings.Cut(d.Id(), "/")

	// if dataset separator not found, we will assume its the bare alias
	// if thats the case, we need to reassign values since strings.Cut would return (alias "", false)
	if !found {
		alias = dataset
	} else {
		d.Set("dataset", dataset)
	}

	d.Set("alias", alias)

	return []*schema.ResourceData{d}, nil
}

func resourceDerivedColumnCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := getConfiguredClient(meta)
	if err != nil {
		return diagFromErr(err)
	}

	dataset := getDatasetOrAll(d)
	derivedColumn := readDerivedColumn(d)

	derivedColumn, err = client.DerivedColumns.Create(ctx, dataset, derivedColumn)
	if err != nil {
		return diagFromErr(err)
	}

	d.Set("alias", derivedColumn.Alias)
	return resourceDerivedColumnRead(ctx, d, meta)
}

func resourceDerivedColumnRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := getConfiguredClient(meta)
	if err != nil {
		return diagFromErr(err)
	}

	dataset := getDatasetOrAll(d)

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
	client, err := getConfiguredClient(meta)
	if err != nil {
		return diagFromErr(err)
	}

	dataset := getDatasetOrAll(d)
	derivedColumn := readDerivedColumn(d)

	derivedColumn, err = client.DerivedColumns.Update(ctx, dataset, derivedColumn)
	if err != nil {
		return diagFromErr(err)
	}

	d.Set("alias", derivedColumn.Alias)
	return resourceDerivedColumnRead(ctx, d, meta)
}

func resourceDerivedColumnDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := getConfiguredClient(meta)
	if err != nil {
		return diagFromErr(err)
	}

	dataset := getDatasetOrAll(d)

	err = client.DerivedColumns.Delete(ctx, dataset, d.Id())
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
