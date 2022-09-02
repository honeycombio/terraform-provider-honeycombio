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
			"field": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(ValidDatasetDefinitions(), false),
						},
						"value": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(0, 255),
						},
					},
				},
			},
		},
	}
}

func resourceDatasetDefinitionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	dataset := d.Get("dataset").(string)
	// check that data is empty
	//dd, err = client.DatasetDefinitions.Create(ctx, dataset, dd)
	d.SetId(dataset)
	return resourceDatasetDefinitionRead(ctx, d, meta)
}

func resourceDatasetDefinitionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	dataset := d.Get("dataset").(string)

	dd, err := client.DatasetDefinitions.Get(ctx, dataset)
	if err == honeycombio.ErrNotFound {
		return nil
	} else if err != nil {
		return diag.FromErr(err)
	}

	flattendDatasetDefinition := flattenDatasetDefinition(dd)
	d.Set("field", flattendDatasetDefinition)

	d.SetId(dataset)
	return nil
}

func resourceDatasetDefinitionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	dataset := d.Get("dataset").(string)
	field := d.Get("field").([]interface{}) // list of definition fields

	dd := expandDatasetDefinition(field)

	dd, err := client.DatasetDefinitions.Update(ctx, dataset, dd)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceDatasetDefinitionRead(ctx, d, meta)
}

// Convert to Terraform Format
func flattenDatasetDefinition(dd *honeycombio.DatasetDefinition) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)

	// for each field allowed unpack the values and set

	if dd.DurationMs.Name != "" {
		result = append(result, map[string]interface{}{
			"name":  "duration_ms",
			"value": dd.DurationMs.Name,
		})
	}

	if dd.Error.Name != "" {
		result = append(result, map[string]interface{}{
			"name":  "error",
			"value": dd.Error.Name,
		})
	}

	if dd.Name.Name != "" {
		result = append(result, map[string]interface{}{
			"name":  "name",
			"value": dd.Name.Name,
		})
	}

	if dd.ParentID.Name != "" {
		result = append(result, map[string]interface{}{
			"name":  "parent_id",
			"value": dd.ParentID.Name,
		})
	}

	if dd.Route.Name != "" {
		result = append(result, map[string]interface{}{
			"name":  "route",
			"value": dd.Route.Name,
		})
	}

	if dd.ServiceName.Name != "" {
		result = append(result, map[string]interface{}{
			"name":  "service_name",
			"value": dd.ServiceName.Name,
		})
	}

	if dd.SpanID.Name != "" {
		result = append(result, map[string]interface{}{
			"name":  "span_id",
			"value": dd.SpanID.Name,
		})
	}

	if dd.SpanType.Name != "" {
		result = append(result, map[string]interface{}{
			"name":  "span_kind",
			"value": dd.SpanType.Name,
		})
	}

	if dd.AnnotationType.Name != "" {
		result = append(result, map[string]interface{}{
			"name":  "annotation_type",
			"value": dd.AnnotationType.Name,
		})
	}

	if dd.LinkTraceID.Name != "" {
		result = append(result, map[string]interface{}{
			"name":  "link_trace_id",
			"value": dd.LinkTraceID.Name,
		})
	}

	if dd.LinkSpanID.Name != "" {
		result = append(result, map[string]interface{}{
			"name":  "link_span_id",
			"value": dd.LinkSpanID.Name,
		})
	}

	if dd.Status.Name != "" {
		result = append(result, map[string]interface{}{
			"name":  "status",
			"value": dd.Status.Name,
		})
	}

	if dd.TraceID.Name != "" {
		result = append(result, map[string]interface{}{
			"name":  "trace_id",
			"value": dd.TraceID.Name,
		})
	}

	if dd.User.Name != "" {
		result = append(result, map[string]interface{}{
			"name":  "user",
			"value": dd.User.Name,
		})
	}

	return result
}

// Convert from Terraform to API Schema
func expandDatasetDefinition(s []interface{}) *honeycombio.DatasetDefinition {
	definition := honeycombio.DatasetDefinition{}

	for _, r := range s {
		rMap := r.(map[string]interface{})

		if rMap["name"].(string) == "duration_ms" {
			definition.DurationMs.Name = rMap["value"].(string)
		} else if rMap["name"].(string) == "error" {
			definition.Error.Name = rMap["value"].(string)
		} else if rMap["name"].(string) == "name" {
			definition.Name.Name = rMap["value"].(string)
		} else if rMap["name"].(string) == "name" {
			definition.ParentID.Name = rMap["value"].(string)
		} else if rMap["name"].(string) == "parent_id" {
			definition.Route.Name = rMap["value"].(string)
		} else if rMap["name"].(string) == "route" {
			definition.ServiceName.Name = rMap["value"].(string)
		} else if rMap["name"].(string) == "service_name" {
			definition.SpanID.Name = rMap["value"].(string)
		} else if rMap["name"].(string) == "span_id" {
			definition.SpanType.Name = rMap["value"].(string)
		} else if rMap["name"].(string) == "span_kind" {
			definition.AnnotationType.Name = rMap["value"].(string)
		} else if rMap["name"].(string) == "annotation_type" {
			definition.LinkTraceID.Name = rMap["value"].(string)
		} else if rMap["name"].(string) == "link_trace_id" {
			definition.LinkSpanID.Name = rMap["value"].(string)
		} else if rMap["name"].(string) == "link_span_id" {
			definition.Status.Name = rMap["value"].(string)
		} else if rMap["name"].(string) == "trace_id" {
			definition.TraceID.Name = rMap["value"].(string)
		} else if rMap["name"].(string) == "user" {
			definition.User.Name = rMap["value"].(string)
		}
	}
	return &definition
}

//func matchDatasetDefinitionWithSchema(datasetDefinition *honeycombio.DatasetDefinition, fields []interface{}) *honeycombio.DatasetDefinition {
//result := honeycombio.DatasetDefinition{}

//rMap := make(map[string]honeycombio.DatasetDefinition, 0)

//// check all field
//for _, recipient := range readRecipients {
//rMap[recipient.ID] = recipient
//}

//// Build up result, with each readRecipient in the same position as it
//// appears in declaredRecipients, by looking at each declaredRecipient and
//// finding its matching readRecipient (via rMap).
////
//// If the declaredRecipient has an ID, this is easy: just look it up and
//// put it in it's place. Otherwise, try to match it to a readRecipient with
//// the same type and target. If we can't find it at all, it must be new, so
//// put it at the end.
//for _, field := range fields {
//declaredField := field.(map[string]interface{})

////if declaVredField["name"] != "" {

//if v, ok := rMap[declaredField["name"].(string)]; ok {
//// matched recipient declared by ID
//result = append(result, v)
//delete(rMap, v.ID)
//}
//} else {
//// group result recipients by type
//for key, rcpt := range rMap {
//if string(rcpt.Type) == declaredRcpt["type"] && rcpt.Target == declaredRcpt["target"] {
//result = append(result, rcpt)
//delete(rMap, key)
//break
//}
//}
//}
//}

//// append unmatched read recipients to the result
//for _, rcpt := range rMap {
//result = append(result, rcpt)
//}

//return result
//}
