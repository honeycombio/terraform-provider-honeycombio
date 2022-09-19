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
		DeleteContext: resourceDatasetDefinitionDelete,

		Schema: map[string]*schema.Schema{
			"dataset": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"field": {
				Type:     schema.TypeSet,
				Optional: true,
				MinItems: 1,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringInSlice(ValidDatasetDefinitions(), false),
								validation.StringLenBetween(0, 255),
							),
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
	d.SetId(d.Get("dataset").(string))
	// _if_ someone wants to pass a datasetDefinition object in for create - should we support that? It would
	// simply be an update
	return resourceDatasetDefinitionRead(ctx, d, meta)
}

// resourceDatasetDefinitionRead pulls the dataset defintion settings from Honeycomb and sets the Terraform state
// to match
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
	err = d.Set("field", flattendDatasetDefinition)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(dataset)
	return nil
}

func resourceDatasetDefinitionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)
	dataset := d.Get("dataset").(string)
	field := d.Get("field").(*schema.Set) // list of definition fields

	dd := expandDatasetDefinition(field)

	dd, err := client.DatasetDefinitions.Update(ctx, dataset, dd)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceDatasetDefinitionRead(ctx, d, meta)
}

func resourceDatasetDefinitionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	dataset := d.Get("dataset").(string)

	emptyDatasetDefinition := &honeycombio.DatasetDefinition{
		DurationMs:     honeycombio.DefinitionColumn{Name: ""},
		Error:          honeycombio.DefinitionColumn{Name: ""},
		Name:           honeycombio.DefinitionColumn{Name: ""},
		ParentID:       honeycombio.DefinitionColumn{Name: ""},
		Route:          honeycombio.DefinitionColumn{Name: ""},
		ServiceName:    honeycombio.DefinitionColumn{Name: ""},
		SpanID:         honeycombio.DefinitionColumn{Name: ""},
		SpanType:       honeycombio.DefinitionColumn{Name: ""},
		AnnotationType: honeycombio.DefinitionColumn{Name: ""},
		LinkTraceID:    honeycombio.DefinitionColumn{Name: ""},
		LinkSpanID:     honeycombio.DefinitionColumn{Name: ""},
		Status:         honeycombio.DefinitionColumn{Name: ""},
		TraceID:        honeycombio.DefinitionColumn{Name: ""},
		User:           honeycombio.DefinitionColumn{Name: ""},
	}

	// set each definition to blank:
	flattenedDatasetDefinition := flattenDatasetDefinition(emptyDatasetDefinition)
	err := d.Set("field", flattenedDatasetDefinition)
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.DatasetDefinitions.Delete(ctx, dataset)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return resourceDatasetDefinitionRead(ctx, d, meta)
}

// Convert to Terraform Format
func flattenDatasetDefinition(dd *honeycombio.DatasetDefinition) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)

	// for each field allowed unpack the values and set
	if dd.DurationMs.Name != "" && !CheckDatasetDefinitionDurationMs(dd.DurationMs.Name) {
		result = append(result, map[string]interface{}{
			"name":  "duration_ms",
			"value": dd.DurationMs.Name,
		})
	}

	if dd.Error.Name != "" && !CheckDatasetDefinitionError(dd.Error.Name) {
		result = append(result, map[string]interface{}{
			"name":  "error",
			"value": dd.Error.Name,
		})

	}

	if dd.Name.Name != "" && !CheckDatasetDefinitionName(dd.Name.Name) {
		result = append(result, map[string]interface{}{
			"name":  "name",
			"value": dd.Name.Name,
		})
	}

	if dd.ParentID.Name != "" && !CheckDatasetDefinitionParentID(dd.ParentID.Name) {
		result = append(result, map[string]interface{}{
			"name":  "parent_id",
			"value": dd.ParentID.Name,
		})
	}

	if dd.Route.Name != "" && !CheckDatasetDefinitionName(dd.Name.Name) {
		result = append(result, map[string]interface{}{
			"name":  "route",
			"value": dd.Route.Name,
		})
	}

	if dd.ServiceName.Name != "" && !CheckDatasetDefinitionServiceName(dd.ServiceName.Name) {
		result = append(result, map[string]interface{}{
			"name":  "service_name",
			"value": dd.ServiceName.Name,
		})
	}

	if dd.SpanID.Name != "" && !CheckDatasetDefinitionSpanID(dd.SpanID.Name) {
		result = append(result, map[string]interface{}{
			"name":  "span_id",
			"value": dd.SpanID.Name,
		})
	}

	if dd.SpanType.Name != "" && !CheckDatasetDefinitionSpanType(dd.SpanType.Name) {
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

	if dd.LinkTraceID.Name != "" && !CheckDatasetDefinitionLinkTraceID(dd.LinkTraceID.Name) {
		result = append(result, map[string]interface{}{
			"name":  "link_trace_id",
			"value": dd.LinkTraceID.Name,
		})
	}

	if dd.LinkSpanID.Name != "" && !CheckDatasetDefinitionLinkSpanID(dd.LinkSpanID.Name) {
		result = append(result, map[string]interface{}{
			"name":  "link_span_id",
			"value": dd.LinkSpanID.Name,
		})
	}

	if dd.Status.Name != "" && !CheckDatasetDefinitionStatus(dd.Status.Name) {
		result = append(result, map[string]interface{}{
			"name":  "status",
			"value": dd.Status.Name,
		})
	}

	if dd.TraceID.Name != "" && !CheckDatasetDefinitionTraceID(dd.TraceID.Name) {
		result = append(result, map[string]interface{}{
			"name":  "trace_id",
			"value": dd.TraceID.Name,
		})
	}

	if dd.User.Name != "" && !CheckDatasetDefinitionUser(dd.User.Name) {
		result = append(result, map[string]interface{}{
			"name":  "user",
			"value": dd.User.Name,
		})
	}

	return result
}

// Convert from Terraform to API Schema
func expandDatasetDefinition(s *schema.Set) *honeycombio.DatasetDefinition {
	definition := honeycombio.DatasetDefinition{}

	for _, r := range s.List() {
		rMap := r.(map[string]interface{})

		if rMap["name"].(string) == "duration_ms" {
			definition.DurationMs.Name = rMap["value"].(string)
		} else if rMap["name"].(string) == "error" {
			definition.Error.Name = rMap["value"].(string)
		} else if rMap["name"].(string) == "name" {
			definition.Name.Name = rMap["value"].(string)
		} else if rMap["name"].(string) == "parent_id" {
			definition.ParentID.Name = rMap["value"].(string)
		} else if rMap["name"].(string) == "route" {
			definition.Route.Name = rMap["value"].(string)
		} else if rMap["name"].(string) == "service_name" {
			definition.ServiceName.Name = rMap["value"].(string)
		} else if rMap["name"].(string) == "span_id" {
			definition.SpanID.Name = rMap["value"].(string)
		} else if rMap["name"].(string) == "span_kind" {
			definition.SpanType.Name = rMap["value"].(string)
		} else if rMap["name"].(string) == "annotation_type" {
			definition.AnnotationType.Name = rMap["value"].(string)
		} else if rMap["name"].(string) == "link_trace_id" {
			definition.LinkTraceID.Name = rMap["value"].(string)
		} else if rMap["name"].(string) == "link_span_id" {
			definition.LinkSpanID.Name = rMap["value"].(string)
		} else if rMap["name"].(string) == "status" {
			definition.Status.Name = rMap["value"].(string)
		} else if rMap["name"].(string) == "trace_id" {
			definition.TraceID.Name = rMap["value"].(string)
		} else if rMap["name"].(string) == "user" {
			definition.User.Name = rMap["value"].(string)
		}
	}
	return &definition
}
