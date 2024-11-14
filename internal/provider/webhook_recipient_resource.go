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
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/validation"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/models"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &webhookRecipientResource{}
	_ resource.ResourceWithConfigure   = &webhookRecipientResource{}
	_ resource.ResourceWithImportState = &webhookRecipientResource{}
)

type webhookRecipientResource struct {
	client *client.Client
}

func NewWebhookRecipientResource() resource.Resource {
	return &webhookRecipientResource{}
}

func (*webhookRecipientResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_webhook_recipient"
}

func (r *webhookRecipientResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	w := getClientFromResourceRequest(&req)
	if w == nil {
		return
	}

	c, err := w.V1Client()
	if err != nil || c == nil {
		resp.Diagnostics.AddError("Failed to configure client", err.Error())
		return
	}
	r.client = c
}

func (*webhookRecipientResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A Webhook recipient can be used by Triggers or BurnAlerts notifications to send an event to an HTTP endpoint.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier for this Recipient.",
				Computed:    true,
				Required:    false,
				Optional:    false,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of this Webhook recipient.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
			},
			"secret": schema.StringAttribute{
				Description: "The secret to include when sending the notification to the webhook.",
				Optional:    true,
				Sensitive:   true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(255),
				},
			},
			"url": schema.StringAttribute{
				Description: "The URL of the endpoint the notification will be sent to.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 2048),
					validation.IsURLWithHTTPorHTTPS(),
				},
			},
		},
	}
}

func (r *webhookRecipientResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID == "" {
		resp.Diagnostics.AddError("Invalid Import ID", "The Recipient ID must be provided")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &models.WebhookRecipientModel{
		ID: types.StringValue(req.ID),
	})...)
}

func (r *webhookRecipientResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan models.WebhookRecipientModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rcpt, err := r.client.Recipients.Create(ctx, &client.Recipient{
		Type: client.RecipientTypeWebhook,
		Details: client.RecipientDetails{
			WebhookName:   plan.Name.ValueString(),
			WebhookURL:    plan.URL.ValueString(),
			WebhookSecret: plan.Secret.ValueString(),
		},
	})
	if helper.AddDiagnosticOnError(&resp.Diagnostics, "Creating Honeycomb Webhook Recipient", err) {
		return
	}

	var state models.WebhookRecipientModel
	state.ID = types.StringValue(rcpt.ID)
	state.Name = types.StringValue(rcpt.Details.WebhookName)
	state.URL = types.StringValue(rcpt.Details.WebhookURL)
	if rcpt.Details.WebhookSecret != "" {
		state.Secret = types.StringValue(rcpt.Details.WebhookSecret)
	} else {
		state.Secret = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *webhookRecipientResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.WebhookRecipientModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var detailedErr client.DetailedError
	rcpt, err := r.client.Recipients.Get(ctx, state.ID.ValueString())
	if errors.As(err, &detailedErr) {
		if detailedErr.IsNotFound() {
			// if not found consider it deleted -- so just remove it from state
			resp.State.RemoveResource(ctx)
			return
		} else {
			resp.Diagnostics.Append(helper.NewDetailedErrorDiagnostic(
				"Error Reading Honeycomb Webhook Recipient",
				&detailedErr,
			))
		}
	} else if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Honeycomb Webhook Recipient",
			"Unexpected error reading Webhook Recipient "+state.ID.ValueString()+": "+err.Error(),
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}
	if rcpt.Type != client.RecipientTypeWebhook {
		resp.Diagnostics.AddError(
			"Error Reading Honeycomb Webhook Recipient",
			"Unexpected recipient type "+rcpt.Type.String(),
		)
		return
	}

	state.ID = types.StringValue(rcpt.ID)
	state.Name = types.StringValue(rcpt.Details.WebhookName)
	state.URL = types.StringValue(rcpt.Details.WebhookURL)
	if rcpt.Details.WebhookSecret != "" {
		state.Secret = types.StringValue(rcpt.Details.WebhookSecret)
	} else {
		state.Secret = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *webhookRecipientResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan models.WebhookRecipientModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Recipients.Update(ctx, &client.Recipient{
		ID:   plan.ID.ValueString(),
		Type: client.RecipientTypeWebhook,
		Details: client.RecipientDetails{
			WebhookName:   plan.Name.ValueString(),
			WebhookURL:    plan.URL.ValueString(),
			WebhookSecret: plan.Secret.ValueString(),
		},
	})
	if helper.AddDiagnosticOnError(&resp.Diagnostics, "Updating Honeycomb Webhook Recipient", err) {
		return
	}

	rcpt, err := r.client.Recipients.Get(ctx, plan.ID.ValueString())
	if helper.AddDiagnosticOnError(&resp.Diagnostics, "Updating Honeycomb Burn Alert", err) {
		return
	}

	var state models.WebhookRecipientModel
	state.ID = types.StringValue(rcpt.ID)
	state.Name = types.StringValue(rcpt.Details.WebhookName)
	state.URL = types.StringValue(rcpt.Details.WebhookURL)
	if rcpt.Details.WebhookSecret != "" {
		state.Secret = types.StringValue(rcpt.Details.WebhookSecret)
	} else {
		state.Secret = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *webhookRecipientResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state models.WebhookRecipientModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var detailedErr client.DetailedError
	err := r.client.Recipients.Delete(ctx, state.ID.ValueString())
	if err != nil {
		if errors.As(err, &detailedErr) {
			// if not found consider it deleted -- so don't error
			if !detailedErr.IsNotFound() {
				resp.Diagnostics.Append(helper.NewDetailedErrorDiagnostic(
					"Error Deleting Honeycomb Webhook Recipient",
					&detailedErr,
				))
			}
		} else {
			resp.Diagnostics.AddError(
				"Error Deleting Honeycomb Webhook Recipient",
				"Could not delete Webhook Recipient ID "+state.ID.ValueString()+": "+err.Error(),
			)
		}
	}
}