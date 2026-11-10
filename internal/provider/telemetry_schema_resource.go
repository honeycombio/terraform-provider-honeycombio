package provider

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
	v2client "github.com/honeycombio/terraform-provider-honeycombio/client/v2"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/validation"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/models"
)

// maxTelemetrySchemaSize is the largest schema, in bytes, which the API accepts.
const maxTelemetrySchemaSize = 1 << 20

// Ensure the implementation satisfies the expected interfaces.
//
// This resource is not implementing ResourceWithImportState:
// it is not importable to start.
var (
	_ resource.Resource              = &telemetrySchemaResource{}
	_ resource.ResourceWithConfigure = &telemetrySchemaResource{}
)

type telemetrySchemaResource struct {
	client *v2client.Client
}

func NewTelemetrySchemaResource() resource.Resource {
	return &telemetrySchemaResource{}
}

func (*telemetrySchemaResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_telemetry_schema"
}

func (r *telemetrySchemaResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	w := getClientFromResourceRequest(&req)
	if w == nil {
		return
	}

	c, err := w.V2Client()
	if err != nil {
		resp.Diagnostics.AddError("Failed to configure client", err.Error())
		return
	}
	r.client = c
}

func (*telemetrySchemaResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the custom telemetry schema of a Honeycomb Team.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				Required:            false,
				Optional:            false,
				MarkdownDescription: "The slug of the Team the schema belongs to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"schema_yaml": schema.StringAttribute{
				Required: true,
				MarkdownDescription: "The telemetry schema, as the YAML of a [Weaver](https://github.com/open-telemetry/weaver) registry, " +
					"of up to 1 MiB (1,048,576 bytes). " +
					"It is stored and returned exactly as written, so any change to the text, including whitespace and comments, is a change to the schema.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					validation.StringByteLengthAtMost(maxTelemetrySchemaSize),
				},
			},
		},
	}
}

func (r *telemetrySchemaResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan models.TelemetrySchemaResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// The Team has a single schema and setting it replaces whatever is there,
	// so look first rather than overwrite a schema Terraform doesn't know about.
	var detailedErr client.DetailedError
	existing, err := r.client.TelemetrySchemas.Get(ctx)
	switch {
	case err == nil:
		if existing.SchemaYAML != plan.SchemaYAML.ValueString() {
			resp.Diagnostics.AddError(
				"Honeycomb Telemetry Schema Already Exists",
				"This Team already has a telemetry schema which differs from the configuration, and creating this resource would replace it. "+
					"To manage the existing schema with Terraform, set `schema_yaml` to match it exactly. "+
					"To replace it, delete the existing schema first.",
			)
			return
		}
		// the existing schema is exactly what was asked for, so adopt it
		resp.Diagnostics.Append(resp.State.Set(ctx, flattenTelemetrySchema(existing))...)
		return
	case errors.As(err, &detailedErr) && detailedErr.IsNotFound():
		// no schema yet, so create it below
	default:
		helper.AddDiagnosticOnError(&resp.Diagnostics, "Creating Honeycomb Telemetry Schema", err)
		return
	}

	schema, err := r.client.TelemetrySchemas.Put(ctx, &v2client.TelemetrySchema{
		SchemaYAML: plan.SchemaYAML.ValueString(),
	})
	if helper.AddDiagnosticOnError(&resp.Diagnostics, "Creating Honeycomb Telemetry Schema", err) {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, flattenTelemetrySchema(schema))...)
}

func (r *telemetrySchemaResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.TelemetrySchemaResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var detailedErr client.DetailedError
	schema, err := r.client.TelemetrySchemas.Get(ctx)
	if errors.As(err, &detailedErr) {
		if detailedErr.IsNotFound() {
			// if not found consider it deleted -- so just remove it from state
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.Append(helper.NewDetailedErrorDiagnostic(
			"Error Reading Honeycomb Telemetry Schema",
			&detailedErr,
		))
	} else if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Honeycomb Telemetry Schema",
			"Unexpected error reading the Telemetry Schema: "+err.Error(),
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, flattenTelemetrySchema(schema))...)
}

func (r *telemetrySchemaResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan models.TelemetrySchemaResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	schema, err := r.client.TelemetrySchemas.Put(ctx, &v2client.TelemetrySchema{
		SchemaYAML: plan.SchemaYAML.ValueString(),
	})
	if helper.AddDiagnosticOnError(&resp.Diagnostics, "Updating Honeycomb Telemetry Schema", err) {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, flattenTelemetrySchema(schema))...)
}

func (r *telemetrySchemaResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	err := r.client.TelemetrySchemas.Delete(ctx)
	var detailedErr client.DetailedError
	if err != nil {
		if errors.As(err, &detailedErr) {
			// if not found consider it deleted -- so don't error
			if !detailedErr.IsNotFound() {
				resp.Diagnostics.Append(helper.NewDetailedErrorDiagnostic(
					"Error Deleting Honeycomb Telemetry Schema",
					&detailedErr,
				))
			}
		} else {
			resp.Diagnostics.AddError(
				"Error Deleting Honeycomb Telemetry Schema",
				"Could not delete the Telemetry Schema: "+err.Error(),
			)
		}
	}
}

// flattenTelemetrySchema copies the schema document into the model untouched.
// The text must never be normalised, or Terraform would report permanent drift.
func flattenTelemetrySchema(s *v2client.TelemetrySchema) models.TelemetrySchemaResourceModel {
	return models.TelemetrySchemaResourceModel{
		ID:         types.StringValue(s.ID),
		SchemaYAML: types.StringValue(s.SchemaYAML),
	}
}
