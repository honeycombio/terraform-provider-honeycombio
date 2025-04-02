package honeycombio

import (
	"context"
	"errors"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper"
)

func newSLO() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSLOCreate,
		ReadContext:   resourceSLORead,
		UpdateContext: resourceSLOUpdate,
		DeleteContext: resourceSLODelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceSLOImport,
		},
		Description: "Honeycomb SLOs allows you to define and monitor Service Level Objectives (SLOs) for your organization.",

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 120),
				Description:  "The name of the SLO.",
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1023),
				Description:  "A description of the SLO's intent and context.",
			},
			"dataset": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The dataset this SLO is created in. Must be the same dataset as the SLI unless the SLI's dataset is `\"__all__\"`.",
				DiffSuppressFunc: func(k, oldValue, newValue string, d *schema.ResourceData) bool {
					if oldValue == newValue {
						return true
					}
					// if the config moves away from deprecated dataset, nothing should change
					if newValue == "" {
						return true
					}
					return false
				},
				ExactlyOneOf: []string{"dataset", "datasets"},
			},
			"datasets": {
				Type:         schema.TypeSet,
				Elem:         &schema.Schema{Type: schema.TypeString},
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				Description:  "The datasets the SLO is evaluated on.",
				ExactlyOneOf: []string{"dataset", "datasets"},
				MaxItems:     10,
				MinItems:     1,
			},
			"sli": {
				Type:     schema.TypeString,
				Required: true,
				Description: `The alias of the Derived Column that will be used as the SLI to indicate event success.
The derived column used as the SLI must be in the same dataset as the SLO. Additionally,
the column evaluation should consistently return nil, true, or false, as these are the only valid values for an SLI.`,
			},
			"target_percentage": {
				Type:         schema.TypeFloat,
				Required:     true,
				Description:  "The percentage of qualified events that you expect to succeed during the `time_period`.",
				ValidateFunc: validation.FloatBetween(0.00000, 99.9999),
			},
			"time_period": {
				Type:         schema.TypeInt,
				Required:     true,
				Description:  "The time period, in days, over which your SLO will be evaluated.",
				ValidateFunc: validation.IntAtLeast(1),
			},
		},
	}
}

func resourceSLOImport(ctx context.Context, d *schema.ResourceData, i interface{}) ([]*schema.ResourceData, error) {
	dataset, id, found := strings.Cut(d.Id(), "/")

	// if dataset separator not found, we will assume its the bare id
	// if thats the case, we need to reassign values since strings.Cut would return (id, "", false)
	if !found {
		id = dataset
	} else {
		d.Set("dataset", dataset)
	}

	d.SetId(id)

	return []*schema.ResourceData{d}, nil
}

func resourceSLOCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := getConfiguredClient(meta)
	if err != nil {
		return diagFromErr(err)
	}

	dataset := getDataset(d)

	s, err := client.SLOs.Create(ctx, dataset, expandSLO(d))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(s.ID)
	return resourceSLORead(ctx, d, meta)
}

func resourceSLORead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := getConfiguredClient(meta)
	if err != nil {
		return diagFromErr(err)
	}

	dataset := getDataset(d)

	var detailedErr honeycombio.DetailedError
	s, err := client.SLOs.Get(ctx, dataset, d.Id())
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

	d.SetId(s.ID)
	d.Set("name", s.Name)
	d.Set("description", s.Description)
	d.Set("sli", s.SLI.Alias)
	d.Set("target_percentage", helper.PPMToFloat(s.TargetPerMillion))
	d.Set("time_period", s.TimePeriodDays)
	d.Set("datasets", s.DatasetSlugs)

	return nil
}

func resourceSLOUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := getConfiguredClient(meta)
	if err != nil {
		return diagFromErr(err)
	}

	dataset := getDataset(d)

	s, err := client.SLOs.Update(ctx, dataset, expandSLO(d))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(s.ID)
	return resourceSLORead(ctx, d, meta)
}

func resourceSLODelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := getConfiguredClient(meta)
	if err != nil {
		return diagFromErr(err)
	}

	dataset := getDataset(d)

	err = client.SLOs.Delete(ctx, dataset, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func expandSLO(d *schema.ResourceData) *honeycombio.SLO {
	var datasets []string
	if v, ok := d.GetOk("datasets"); ok {
		for _, v := range v.(*schema.Set).List() {
			datasets = append(datasets, v.(string))
		}
	}

	return &honeycombio.SLO{
		ID:               d.Id(),
		Name:             d.Get("name").(string),
		Description:      d.Get("description").(string),
		TimePeriodDays:   d.Get("time_period").(int),
		TargetPerMillion: helper.FloatToPPM(d.Get("target_percentage").(float64)),
		SLI:              honeycombio.SLIRef{Alias: d.Get("sli").(string)},
		DatasetSlugs:     datasets,
	}
}

func getDataset(d *schema.ResourceData) string {
	dataset := d.Get("dataset").(string)
	if dataset == "" {
		dataset = "__all__"
	}
	return dataset
}
