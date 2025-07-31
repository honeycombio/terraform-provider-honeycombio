package honeycombio

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/honeycombio/internal/verify"
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
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Description:      "The dataset this query annotation is added to. If not set, an Environment-wide query annotation will be created.",
				DiffSuppressFunc: verify.SuppressEquivEnvWideDataset,
			},
			"query_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the query that the annotation will be created on. Note that a query can have more than one annotation.",
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The name to display as the query annotation.",
				ValidateFunc: validation.StringLenBetween(1, 320),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "The description to display as the query annotation.",
				ValidateFunc: validation.StringLenBetween(1, 1023),
			},
		},
	}
}

func buildQueryAnnotation(d *schema.ResourceData) *honeycombio.QueryAnnotation {
	name, ok := d.Get("name").(string)
	if !ok {
		panic("name must be a string")
	}
	description, ok := d.Get("description").(string)
	if !ok {
		panic("description must be a string")
	}
	queryID, ok := d.Get("query_id").(string)
	if !ok {
		panic("query_id must be a string")
	}
	return &honeycombio.QueryAnnotation{
		ID:          d.Id(),
		Name:        name,
		Description: description,
		QueryID:     queryID,
	}
}

func resourceQueryAnnotationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := getConfiguredClient(meta)
	if err != nil {
		return diagFromErr(err)
	}
	dataset := getDatasetOrAll(d)
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
	dataset := getDatasetOrAll(d)
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
	dataset, ok := d.Get("dataset").(string)
	if !ok {
		return diag.Errorf("dataset must be a string")
	}

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
	dataset := getDatasetOrAll(d)

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

	_ = d.Set("name", queryAnnotation.Name)
	_ = d.Set("description", queryAnnotation.Description)
	_ = d.Set("query_id", queryAnnotation.QueryID)
	return nil
}
