package honeycombio

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/honeycombio/internal/hashcode"
)

func newDatasetDefinition() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDatasetDefinitionUpdate, // datasets implicitly have definitions already, so we're always updating
		ReadContext:   resourceDatasetDefinitionRead,
		UpdateContext: resourceDatasetDefinitionUpdate,
		DeleteContext: resourceDatasetDefinitionDelete,

		Schema: map[string]*schema.Schema{
			"dataset": {
				Type:        schema.TypeString,
				Description: "The dataset to set the Dataset Definition for.",
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Type:         schema.TypeString,
				Description:  " The name of the definition being set.",
				Required:     true,
				ValidateFunc: validation.StringInSlice(honeycombio.DatasetDefinitionColumns(), false),
			},
			"column": {
				Type:         schema.TypeString,
				Description:  "The column to set the definition to. Must be the name of an existing Column or the alias of an existing Derived Column.",
				Required:     true,
				ValidateFunc: validation.StringLenBetween(0, 255),
			},
		},
	}
}

func resourceDatasetDefinitionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	dataset := d.Get("dataset").(string)
	dd, err := client.DatasetDefinitions.Get(ctx, dataset)
	if err == honeycombio.ErrNotFound {
		d.SetId("")
		return nil
	} else if err != nil {
		return diag.FromErr(err)
	}

	name := d.Get("name").(string)
	column := extractDatasetDefinitionByName(name, dd)

	d.SetId(strconv.Itoa(hashcode.String(fmt.Sprintf("%s-%s", dataset, name))))
	d.Set("name", name)
	d.Set("column", column)

	return nil
}

func resourceDatasetDefinitionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	dataset := d.Get("dataset").(string)
	name := d.Get("name").(string)
	value := d.Get("column").(string)

	dd := expandDatasetDefinition(name, value)
	_, err := client.DatasetDefinitions.Update(ctx, dataset, dd)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceDatasetDefinitionRead(ctx, d, meta)
}

func resourceDatasetDefinitionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	dataset := d.Get("dataset").(string)
	name := d.Get("name").(string)

	// 'deleting' a definition is really resetting it
	dd := expandDatasetDefinition(name, "")
	_, err := client.DatasetDefinitions.Update(ctx, dataset, dd)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func extractDatasetDefinitionByName(name string, dd *honeycombio.DatasetDefinition) string {
	switch name {
	case "duration_ms":
		return dd.DurationMs.Name
	case "error":
		return dd.Error.Name
	case "name":
		return dd.Name.Name
	case "parent_id":
		return dd.ParentID.Name
	case "route":
		return dd.Route.Name
	case "service_name":
		return dd.ServiceName.Name
	case "span_id":
		return dd.SpanID.Name
	case "span_kind":
		return dd.SpanKind.Name
	case "annotation_type":
		return dd.AnnotationType.Name
	case "link_trace_id":
		return dd.LinkTraceID.Name
	case "link_span_id":
		return dd.LinkSpanID.Name
	case "status":
		return dd.Status.Name
	case "trace_id":
		return dd.TraceID.Name
	case "user":
		return dd.User.Name
	default:
		return ""
	}
}

func expandDatasetDefinition(name, value string) *honeycombio.DatasetDefinition {
	definition := &honeycombio.DatasetDefinition{}

	switch name {
	case "duration_ms":
		definition.DurationMs = &honeycombio.DefinitionColumn{Name: value}
	case "error":
		definition.Error = &honeycombio.DefinitionColumn{Name: value}
	case "name":
		definition.Name = &honeycombio.DefinitionColumn{Name: value}
	case "parent_id":
		definition.ParentID = &honeycombio.DefinitionColumn{Name: value}
	case "route":
		definition.Route = &honeycombio.DefinitionColumn{Name: value}
	case "service_name":
		definition.ServiceName = &honeycombio.DefinitionColumn{Name: value}
	case "span_id":
		definition.SpanID = &honeycombio.DefinitionColumn{Name: value}
	case "span_kind":
		definition.SpanKind = &honeycombio.DefinitionColumn{Name: value}
	case "annotation_type":
		definition.AnnotationType = &honeycombio.DefinitionColumn{Name: value}
	case "link_trace_id":
		definition.LinkTraceID = &honeycombio.DefinitionColumn{Name: value}
	case "link_span_id":
		definition.LinkSpanID = &honeycombio.DefinitionColumn{Name: value}
	case "status":
		definition.Status = &honeycombio.DefinitionColumn{Name: value}
	case "trace_id":
		definition.TraceID = &honeycombio.DefinitionColumn{Name: value}
	case "user":
		definition.User = &honeycombio.DefinitionColumn{Name: value}
	}

	return definition
}
