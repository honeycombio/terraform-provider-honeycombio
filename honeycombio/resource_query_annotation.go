package honeycombio

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
)

func newQueryAnnotation() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceQueryAnnotationCreate,
		ReadContext:   resourceQueryAnnotationRead,
		UpdateContext: resourceQueryAnnotationUpdate,
		DeleteContext: resourceQueryAnnotationDestroy,
		Importer:      nil,

		Schema: map[string]*schema.Schema{
			"dataset": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"query_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 120),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 280),
			},
		},
	}
}

func buildQueryAnnotation(d *schema.ResourceData) *honeycombio.QueryAnnotation {
	return &honeycombio.QueryAnnotation{
		ID:          d.Id(),
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		QueryID:     d.Get("query_id").(string),
	}
}

func resourceQueryAnnotationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)
	dataset := d.Get("dataset").(string)
	queryAnnotation := buildQueryAnnotation(d)

	queryAnnotation, err := client.QueryAnnotations.Create(ctx, dataset, queryAnnotation)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(queryAnnotation.ID)
	return resourceQueryAnnotationRead(ctx, d, meta)
}

func resourceQueryAnnotationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)
	dataset := d.Get("dataset").(string)
	queryAnnotation := buildQueryAnnotation(d)

	_, err := client.QueryAnnotations.Update(ctx, dataset, queryAnnotation)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceQueryAnnotationRead(ctx, d, meta)
}

func resourceQueryAnnotationDestroy(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)
	dataset := d.Get("dataset").(string)

	err := client.QueryAnnotations.Delete(ctx, dataset, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceQueryAnnotationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)
	dataset := d.Get("dataset").(string)

	queryAnnotation, err := client.QueryAnnotations.Get(ctx, dataset, d.Id())
	if err != nil {
		if err == honeycombio.ErrNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.Set("name", queryAnnotation.Name)
	d.Set("description", queryAnnotation.Description)
	d.Set("query_id", queryAnnotation.QueryID)
	return nil
}
