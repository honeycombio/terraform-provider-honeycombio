package honeycombio

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
)

func newDatasetDefinition() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDatasetDefinitionCreate,
		ReadContext:   resourceDatasetDefinitionRead,
		UpdateContext: resourceDatasetDefinitionUpdate,
		DeleteContext: schema.NoopContext,

		Schema: map[string]*schema.Schema{
			"dataset": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"trace_id": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 255),
						},
						"column_type": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(ValidDatasetDefinitionsColumnTypes(), false),
						},
					},
				},
			},
		},
	}
}

func resourceDatasetDefinitionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	//client := meta.(*honeycombio.Client)

	dataset := d.Get("dataset").(string)

	//dd, err = client.DatasetDefinitions.Create(ctx, dataset, dd)

	d.SetId(dataset)

	return resourceDatasetDefinitionRead(ctx, d, meta)
}

func resourceDatasetDefinitionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	dataset := d.Get("dataset").(string)

	dd, err := client.DatasetDefinitions.List(ctx, dataset)
	if err == honeycombio.ErrNotFound {
		return nil
	} else if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(dataset)

	d.Set("trace_id", flattenDefinitionColumn(&dd.TraceID))

	return nil
}

func resourceDatasetDefinitionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	dataset := d.Get("dataset").(string)

	dd, err := expandDatasetDefinition(d)
	if err != nil {
		return diag.FromErr(err)
	}

	dd, err = client.DatasetDefinitions.Update(ctx, dataset, dd)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceDatasetDefinitionRead(ctx, d, meta)
}

func expandDatasetDefinition(d *schema.ResourceData) (*honeycombio.DatasetDefinition, error) {
	// expand into individual definition columns
	traceID := honeycombio.DefinitionColumn{
		Name:       d.Get("trace_id").(string),
		ColumnType: d.Get("column_type").(string),
	}

	// expand into Honeycomb Dataset Definition struct
	datasetDefinition := &honeycombio.DatasetDefinition{
		TraceID: traceID,
	}

	return datasetDefinition, nil
}

func flattenDefinitionColumn(dc *honeycombio.DefinitionColumn) []map[string]interface{} {
	return []map[string]interface{}{
		{
			"name":        dc.Name,
			"column_type": dc.ColumnType,
		},
	}
}

func flattenDatasetDefinition(dd *honeycombio.DatasetDefinition) []map[string]interface{} {
	return []map[string]interface{}{
		{
			"trace_id": flattenDefinitionColumn(&dd.TraceID),
		},
	}
}
