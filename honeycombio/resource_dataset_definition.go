package honeycombio

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
	hnyerr "github.com/honeycombio/terraform-provider-honeycombio/client/errors"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/hashcode"
)

func newDatasetDefinition() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDatasetDefinitionUpdate, // datasets implicitly have definitions already, so we're always updating
		ReadContext:   resourceDatasetDefinitionRead,
		UpdateContext: resourceDatasetDefinitionUpdate,
		DeleteContext: resourceDatasetDefinitionDelete,
		Importer:      nil,

		Schema: map[string]*schema.Schema{
			"dataset": {
				Type:        schema.TypeString,
				Description: "The dataset to set the Dataset Definition for.",
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Type:         schema.TypeString,
				Description:  "The name of the definition being set.",
				Required:     true,
				ValidateFunc: validation.StringInSlice(honeycombio.DatasetDefinitionFields(), false),
			},
			"column": {
				Type:         schema.TypeString,
				Description:  "The column to set the definition to. Must be the name of an existing Column or the alias of an existing Derived Column.",
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"column_type": {
				Type:        schema.TypeString,
				Description: "The type of the column assigned to the definition. Either `column` or `derived_column`.",
				Required:    false,
				Optional:    false,
				Computed:    true,
			},
		},
	}
}

func resourceDatasetDefinitionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := getConfiguredClient(meta)
	if err != nil {
		return diagFromErr(err)
	}

	dataset := d.Get("dataset").(string)

	var detailedErr hnyerr.DetailedError
	dd, err := client.DatasetDefinitions.Get(ctx, dataset)
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

	name := d.Get("name").(string)
	column := extractDatasetDefinitionColumnByName(dd, name)

	d.SetId(strconv.Itoa(hashcode.String(fmt.Sprintf("%s-%s", dataset, name))))
	d.Set("name", name)
	d.Set("column", column.Name)
	d.Set("column_type", column.ColumnType)

	return nil
}

func resourceDatasetDefinitionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := getConfiguredClient(meta)
	if err != nil {
		return diagFromErr(err)
	}

	dataset := d.Get("dataset").(string)
	name := d.Get("name").(string)
	value := d.Get("column").(string)

	dd := expandDatasetDefinition(name, value)
	_, err = client.DatasetDefinitions.Update(ctx, dataset, dd)
	if err != nil {
		return diagFromErr(err)
	}

	return resourceDatasetDefinitionRead(ctx, d, meta)
}

func resourceDatasetDefinitionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := getConfiguredClient(meta)
	if err != nil {
		return diagFromErr(err)
	}

	dataset := d.Get("dataset").(string)
	name := d.Get("name").(string)

	// 'deleting' a definition is really resetting it
	dd := expandDatasetDefinition(name, "")
	_, err = client.DatasetDefinitions.Update(ctx, dataset, dd)
	if err != nil {
		return diagFromErr(err)
	}
	return nil
}

func extractDatasetDefinitionColumnByName(dd *honeycombio.DatasetDefinition, name string) *honeycombio.DefinitionColumn {
	switch name {
	case "duration_ms":
		return dd.DurationMs
	case "error":
		return dd.Error
	case "name":
		return dd.Name
	case "parent_id":
		return dd.ParentID
	case "route":
		return dd.Route
	case "service_name":
		return dd.ServiceName
	case "span_id":
		return dd.SpanID
	case "span_kind":
		return dd.SpanKind
	case "annotation_type":
		return dd.AnnotationType
	case "link_trace_id":
		return dd.LinkTraceID
	case "link_span_id":
		return dd.LinkSpanID
	case "status":
		return dd.Status
	case "trace_id":
		return dd.TraceID
	case "user":
		return dd.User
	default:
		return nil
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
