package honeycombio

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
)

func newBurnAlert() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceBurnAlertCreate,
		ReadContext:   resourceBurnAlertRead,
		UpdateContext: resourceBurnAlertUpdate,
		DeleteContext: resourceBurnAlertDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceBurnAlertImport,
		},
		Description: "Burn Alerts are used to notify you when your error budget will be exhausted within a given time period.",

		Schema: map[string]*schema.Schema{
			"exhaustion_minutes": {
				Type:         schema.TypeInt,
				Required:     true,
				Description:  "The amount of time, in minutes, remaining before the SLO's error budget will be exhausted and the alert will fire.",
				ValidateFunc: validation.IntAtLeast(0),
			},
			"slo_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the SLO that this Burn Alert is for.",
			},
			"dataset": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The dataset this Burn Alert is added to. This must be the same as the SLO's dataset.",
			},
			"recipient": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "A Recipient to notify when an alert fires",
				Elem: &schema.Resource{
					// TODO can we validate either id or type+target is set?
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "The ID of an already existing recipient. Should not be used in combination with `type` and `target`.",
						},
						"type": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringInSlice(recipientTypeStrings(), false),
							Description:  "The type of notification. Allowed types are `email`, `pagerduty`, `slack` and `webhook`. Should not be used in combination with `id`.",
						},
						"target": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "Target of the trigger recipient, this has another meaning depending on the type of recipient. Should not be used in combination with `id`.",
						},
					},
				},
			},
		},
	}
}

func resourceBurnAlertImport(ctx context.Context, d *schema.ResourceData, i interface{}) ([]*schema.ResourceData, error) {
	// import ID is of the format <dataset>/<BurnAlert ID>
	// note that the dataset name can also contain '/'
	idSegments := strings.Split(d.Id(), "/")
	if len(idSegments) < 2 {
		return nil, fmt.Errorf("invalid import ID, supplied ID must be written as <dataset>/<BurnAlert ID>")
	}

	dataset := strings.Join(idSegments[0:len(idSegments)-1], "/")
	id := idSegments[len(idSegments)-1]

	d.Set("dataset", dataset)
	d.SetId(id)

	return []*schema.ResourceData{d}, nil
}

func resourceBurnAlertCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	dataset := d.Get("dataset").(string)
	b, err := expandBurnAlert(d)
	if err != nil {
		return diag.FromErr(err)
	}

	b, err = client.BurnAlerts.Create(ctx, dataset, b)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(b.ID)
	return resourceBurnAlertRead(ctx, d, meta)
}

func resourceBurnAlertRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	dataset := d.Get("dataset").(string)

	b, err := client.BurnAlerts.Get(ctx, dataset, d.Id())
	if err != nil {
		if err == honeycombio.ErrNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId(b.ID)
	d.Set("exhaustion_minutes", b.ExhaustionMinutes)
	d.Set("slo_id", b.SLO.ID)

	declaredRecipients, ok := d.Get("recipient").([]interface{})
	if !ok {
		return diag.Errorf("failed to parse recipients for Burn Alert %s", b.ID)
	}
	err = d.Set("recipient", flattenRecipients(matchRecipientsWithSchema(b.Recipients, declaredRecipients)))
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceBurnAlertUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	dataset := d.Get("dataset").(string)
	b, err := expandBurnAlert(d)
	if err != nil {
		return diag.FromErr(err)
	}

	b, err = client.BurnAlerts.Update(ctx, dataset, b)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(b.ID)
	return resourceBurnAlertRead(ctx, d, meta)
}

func resourceBurnAlertDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	dataset := d.Get("dataset").(string)

	err := client.BurnAlerts.Delete(ctx, dataset, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func expandBurnAlert(d *schema.ResourceData) (*honeycombio.BurnAlert, error) {
	b := &honeycombio.BurnAlert{
		ID:                d.Id(),
		ExhaustionMinutes: d.Get("exhaustion_minutes").(int),
		SLO:               honeycombio.SLORef{ID: d.Get("slo_id").(string)},
		Recipients:        expandRecipients(d.Get("recipient").([]interface{})),
	}
	return b, nil
}
