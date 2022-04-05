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

func newTrigger() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTriggerCreate,
		ReadContext:   resourceTriggerRead,
		UpdateContext: resourceTriggerUpdate,
		DeleteContext: resourceTriggerDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceTriggerImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
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
			"disabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"query_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"threshold": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"op": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(triggerThresholdOpStrings(), false),
						},
						"value": {
							Type:     schema.TypeFloat,
							Required: true,
						},
					},
				},
			},
			"frequency": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.All(
					validation.IntDivisibleBy(60),
					validation.IntBetween(60, 86400),
				),
			},
			"recipient": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					// TODO can we validate either id or type+target is set?
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"type": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringInSlice(triggerRecipientTypeStrings(), false),
						},
						"target": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func resourceTriggerImport(ctx context.Context, d *schema.ResourceData, i interface{}) ([]*schema.ResourceData, error) {
	// import ID is of the format <dataset>/<trigger ID>
	// note that the dataset name can also contain '/'
	idSegments := strings.Split(d.Id(), "/")
	if len(idSegments) < 2 {
		return nil, fmt.Errorf("invalid import ID, supplied ID must be written as <dataset>/<trigger ID>")
	}

	dataset := strings.Join(idSegments[0:len(idSegments)-1], "/")
	id := idSegments[len(idSegments)-1]

	d.Set("dataset", dataset)
	d.SetId(id)

	return []*schema.ResourceData{d}, nil
}

func resourceTriggerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	dataset := d.Get("dataset").(string)
	t, err := expandTrigger(d)
	if err != nil {
		return diag.FromErr(err)
	}

	t, err = client.Triggers.Create(ctx, dataset, t)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(t.ID)
	return resourceTriggerRead(ctx, d, meta)
}

func resourceTriggerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	dataset := d.Get("dataset").(string)

	t, err := client.Triggers.Get(ctx, dataset, d.Id())
	if err != nil {
		if err == honeycombio.ErrNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId(t.ID)
	d.Set("name", t.Name)
	d.Set("description", t.Description)
	d.Set("disabled", t.Disabled)
	d.Set("query_id", t.QueryID)

	err = d.Set("threshold", flattenTriggerThreshold(t.Threshold))
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("frequency", t.Frequency)

	declaredRecipients, ok := d.Get("recipient").([]interface{})
	if !ok {
		return diag.Errorf("failed to parse recipients for Trigger %s", t.ID)
	}
	err = d.Set("recipient", flattenTriggerRecipients(matchRecipientsWithSchema(t.Recipients, declaredRecipients)))
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceTriggerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	dataset := d.Get("dataset").(string)
	t, err := expandTrigger(d)
	if err != nil {
		return diag.FromErr(err)
	}

	t, err = client.Triggers.Update(ctx, dataset, t)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(t.ID)
	return resourceTriggerRead(ctx, d, meta)
}

func resourceTriggerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	dataset := d.Get("dataset").(string)

	err := client.Triggers.Delete(ctx, dataset, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func expandTrigger(d *schema.ResourceData) (*honeycombio.Trigger, error) {
	trigger := &honeycombio.Trigger{
		ID:          d.Id(),
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Disabled:    d.Get("disabled").(bool),
		QueryID:     d.Get("query_id").(string),
		Threshold:   expandTriggerThreshold(d.Get("threshold").([]interface{})),
		Frequency:   d.Get("frequency").(int),
		Recipients:  expandTriggerRecipients(d.Get("recipient").([]interface{})),
	}
	return trigger, nil
}

func expandTriggerThreshold(s []interface{}) *honeycombio.TriggerThreshold {
	d := s[0].(map[string]interface{})

	return &honeycombio.TriggerThreshold{
		Op:    honeycombio.TriggerThresholdOp(d["op"].(string)),
		Value: d["value"].(float64),
	}
}

func expandTriggerRecipients(s []interface{}) []honeycombio.TriggerRecipient {
	triggerRecipients := make([]honeycombio.TriggerRecipient, len(s))

	for i, r := range s {
		rMap := r.(map[string]interface{})

		triggerRecipients[i] = honeycombio.TriggerRecipient{
			ID:     rMap["id"].(string),
			Type:   honeycombio.TriggerRecipientType(rMap["type"].(string)),
			Target: rMap["target"].(string),
		}
	}

	return triggerRecipients
}

func flattenTriggerThreshold(t *honeycombio.TriggerThreshold) []map[string]interface{} {
	return []map[string]interface{}{
		{
			"op":    t.Op,
			"value": t.Value,
		},
	}
}

func flattenTriggerRecipients(rs []honeycombio.TriggerRecipient) []map[string]interface{} {
	result := make([]map[string]interface{}, len(rs))

	for i, r := range rs {
		result[i] = map[string]interface{}{
			"id":     r.ID,
			"type":   string(r.Type),
			"target": r.Target,
		}
	}

	return result
}

// Provides a stable order for Trigger recipients.
//
// This cannot currently be handled efficiently by a DiffSuppressFunc.
// See: https://github.com/hashicorp/terraform-plugin-sdk/issues/477
func matchRecipientsWithSchema(recipients []honeycombio.TriggerRecipient, declaredRecipients []interface{}) []honeycombio.TriggerRecipient {
	result := make([]honeycombio.TriggerRecipient, len(declaredRecipients))

	rMap := make(map[string]honeycombio.TriggerRecipient, len(declaredRecipients))
	for _, recipient := range recipients {
		rMap[recipient.ID] = recipient
	}

	for i, declaredRcpt := range declaredRecipients {
		declaredRcpt := declaredRcpt.(map[string]interface{})

		if declaredRcpt["id"] != "" {
			// recipient declared by ID
			if v, ok := rMap[declaredRcpt["id"].(string)]; ok {
				result[i] = v
				delete(rMap, v.ID)
			}
		} else {
			for key, rcpt := range rMap {
				if string(rcpt.Type) != declaredRcpt["type"] || rcpt.Target != declaredRcpt["target"] {
					continue
				}

				result[i] = rcpt
				delete(rMap, key)
				break
			}
		}
	}

	for _, rcpt := range rMap {
		result = append(result, rcpt)
	}

	return result
}
