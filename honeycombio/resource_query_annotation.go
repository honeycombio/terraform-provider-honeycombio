package honeycombio

import (
	"context"
	"errors"

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
				ValidateFunc: validation.StringLenBetween(1, 320),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1023),
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
	client, err := getConfiguredClient(meta)
	if err != nil {
		return diagFromErr(err)
	}
	dataset := d.Get("dataset").(string)
	queryAnnotation := buildQueryAnnotation(d)

	queryAnnotation, err = client.QueryAnnotations.Create(ctx, dataset, queryAnnotation)
	if err != nil {
		return diagFromErr(err)
	}

	d.SetId(queryAnnotation.ID)
	return resourceQueryAnnotationRead(ctx, d, meta)
}

func resourceQueryAnnotationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := getConfiguredClient(meta)
	if err != nil {
		return diagFromErr(err)
	}
	dataset := d.Get("dataset").(string)
	queryAnnotation := buildQueryAnnotation(d)

	_, err = client.QueryAnnotations.Update(ctx, dataset, queryAnnotation)
	if err != nil {
		return diagFromErr(err)
	}

	return resourceQueryAnnotationRead(ctx, d, meta)
}

func resourceQueryAnnotationDestroy(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := getConfiguredClient(meta)
	if err != nil {
		return diagFromErr(err)
	}
	dataset := d.Get("dataset").(string)

	err = client.QueryAnnotations.Delete(ctx, dataset, d.Id())
	if err != nil {
		return diagFromErr(err)
	}
	return nil
}

func resourceQueryAnnotationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := getConfiguredClient(meta)
	if err != nil {
		return diagFromErr(err)
	}
	dataset := d.Get("dataset").(string)

	var detailedErr honeycombio.DetailedError
	queryAnnotation, err := client.QueryAnnotations.Get(ctx, dataset, d.Id())
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

	d.Set("name", queryAnnotation.Name)
	d.Set("description", queryAnnotation.Description)
	d.Set("query_id", queryAnnotation.QueryID)
	return nil
}
