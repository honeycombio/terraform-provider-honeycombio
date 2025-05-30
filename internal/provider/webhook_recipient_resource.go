package provider

import (
	"context"
	"errors"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"golang.org/x/net/http/httpguts"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/validation"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/models"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                   = &webhookRecipientResource{}
	_ resource.ResourceWithConfigure      = &webhookRecipientResource{}
	_ resource.ResourceWithImportState    = &webhookRecipientResource{}
	_ resource.ResourceWithValidateConfig = &webhookRecipientResource{}

	webhookTemplateTypes     = []string{"trigger", "exhaustion_time", "budget_rate"}
	webhookHeaderDefaults    = []string{"Content-Type", "User-Agent", "X-Honeycomb-Webhook-Token"}
	webhookTemplateNameRegex = regexp.MustCompile(`^[a-z](?:[a-zA-Z0-9]+$)?$`)
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
		Blocks: map[string]schema.Block{
			"template": schema.SetNestedBlock{
				Description: "Template for custom webhook payloads",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Required:    true,
							Description: "The type of the webhook template",
							Validators: []validator.String{
								stringvalidator.OneOf(webhookTemplateTypes...),
							},
						},
						"body": schema.StringAttribute{
							Required:    true,
							Description: "JSON formatted string of the webhook payload",
						},
					},
				},
			},
			"variable": schema.SetNestedBlock{
				Description: "Variables for webhook templates",
				Validators: []validator.Set{
					setvalidator.SizeAtMost(10),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:    true,
							Description: "The name of the variable",
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 64),
								stringvalidator.RegexMatches(webhookTemplateNameRegex, "must be an alphanumeric string beginning with a lowercase letter"),
							},
						},
						"default_value": schema.StringAttribute{
							Description: "An optional default value for the variable",
							Optional:    true,
							Computed:    true,
							Default:     stringdefault.StaticString(""),
							Validators: []validator.String{
								stringvalidator.LengthAtMost(256),
							},
						},
					},
				},
			},
			"header": schema.SetNestedBlock{
				Description: "Custom headers for webhooks",
				Validators: []validator.Set{
					setvalidator.SizeAtMost(5),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:    true,
							Description: "The name or key for the header",
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 64),
								stringvalidator.NoneOfCaseInsensitive(webhookHeaderDefaults...),
							},
						},
						"value": schema.StringAttribute{
							Description: "Value for the header",
							Optional:    true,
							Computed:    true,
							Default:     stringdefault.StaticString(""),
							Validators: []validator.String{
								stringvalidator.LengthAtMost(750),
							},
						},
					},
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
		ID:        types.StringValue(req.ID),
		Templates: types.SetUnknown(types.ObjectType{AttrTypes: models.WebhookTemplateAttrType}),
		Variables: types.SetUnknown(types.ObjectType{AttrTypes: models.TemplateVariableAttrType}),
		Headers:   types.SetUnknown(types.ObjectType{AttrTypes: models.WebhookHeaderAttrType}),
	})...)
}

func (r *webhookRecipientResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data models.WebhookRecipientModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var templates []models.WebhookTemplateModel
	data.Templates.ElementsAs(ctx, &templates, false)

	var variables []models.TemplateVariableModel
	data.Variables.ElementsAs(ctx, &variables, false)

	var headers []models.WebhookHeaderModel
	data.Headers.ElementsAs(ctx, &headers, false)

	triggerTmplExists := false
	budgetRateTmplExists := false
	exhaustionTimeTmplExists := false
	for i, t := range templates {
		// only allow one template of each type (trigger, budget_rate, exhaustion_time)
		switch t.Type {
		case types.StringValue("trigger"):
			if triggerTmplExists {
				resp.Diagnostics.AddAttributeError(
					path.Root("template").AtListIndex(i).AtName("type"),
					"Conflicting configuration arguments",
					"cannot have more than one \"template\" of type \"trigger\"",
				)
			}
			triggerTmplExists = true
		case types.StringValue("exhaustion_time"):
			if exhaustionTimeTmplExists {
				resp.Diagnostics.AddAttributeError(
					path.Root("template").AtListIndex(i).AtName("type"),
					"Conflicting configuration arguments",
					"cannot have more than one \"template\" of type \"exhaustion_time\"",
				)
			}
			exhaustionTimeTmplExists = true
		case types.StringValue("budget_rate"):
			if budgetRateTmplExists {
				resp.Diagnostics.AddAttributeError(
					path.Root("template").AtListIndex(i).AtName("type"),
					"Conflicting configuration arguments",
					"cannot have more than one \"template\" of type \"budget_rate\"",
				)
			}
			budgetRateTmplExists = true
		}
	}

	// template variables cannot be configured without a template
	if len(variables) >= 1 && len(templates) == 0 {
		resp.Diagnostics.AddAttributeError(
			path.Root("variable").AtListIndex(0),
			"Conflicting configuration arguments",
			"cannot configure a \"variable\" without also configuring a \"template\"",
		)
	}

	// variable names cannot be duplicated
	duplicateMap := make(map[string]bool)
	for i, v := range variables {
		name := v.Name.ValueString()
		if duplicateMap[name] {
			resp.Diagnostics.AddAttributeError(
				path.Root("variable").AtListIndex(i).AtName("name"),
				"Conflicting configuration arguments",
				"cannot have more than one \"variable\" with the same \"name\"",
			)
		}
		duplicateMap[name] = true
	}

	// webhook headers must be valid http headers
	for i, h := range headers {
		if !httpguts.ValidHeaderFieldName(h.Name.ValueString()) {
			resp.Diagnostics.AddAttributeError(
				path.Root("header").AtListIndex(i).AtName("name"),
				"Conflicting configuration arguments",
				"invalid webhook header name",
			)
		}
		if !httpguts.ValidHeaderFieldValue(h.Value.ValueString()) {
			resp.Diagnostics.AddAttributeError(
				path.Root("header").AtListIndex(i).AtName("value"),
				"Conflicting configuration arguments",
				"invalid webhook header value",
			)
		}
	}
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
			WebhookName:     plan.Name.ValueString(),
			WebhookURL:      plan.URL.ValueString(),
			WebhookSecret:   plan.Secret.ValueString(),
			WebhookPayloads: webhookTemplatesToClientPayloads(ctx, plan.Templates, plan.Variables, &resp.Diagnostics),
			WebhookHeaders:  expandWebhookHeaders(ctx, plan.Headers, &resp.Diagnostics),
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

	// to prevent confusing if/else blocks, set null by default and override it if we have that detail on the recipient
	state.Templates = types.SetNull(types.ObjectType{AttrTypes: models.WebhookTemplateAttrType})
	state.Variables = types.SetNull(types.ObjectType{AttrTypes: models.TemplateVariableAttrType})
	state.Headers = types.SetNull(types.ObjectType{AttrTypes: models.WebhookHeaderAttrType})

	if rcpt.Details.WebhookPayloads != nil {
		state.Templates = plan.Templates
		if rcpt.Details.WebhookPayloads.TemplateVariables != nil {
			state.Variables = plan.Variables
		}
	}

	if rcpt.Details.WebhookHeaders != nil {
		state.Headers = plan.Headers
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

	if rcpt.Details.WebhookPayloads != nil {
		state.Templates, state.Variables = clientPayloadsToWebhookTemplateSets(ctx, rcpt.Details.WebhookPayloads, &resp.Diagnostics)
	} else {
		state.Templates = types.SetNull(types.ObjectType{AttrTypes: models.WebhookTemplateAttrType})
		state.Variables = types.SetNull(types.ObjectType{AttrTypes: models.TemplateVariableAttrType})
	}

	if rcpt.Details.WebhookHeaders != nil {
		state.Headers = flattenWebhookHeaders(ctx, rcpt.Details.WebhookHeaders, &resp.Diagnostics)
	} else {
		state.Headers = types.SetNull(types.ObjectType{AttrTypes: models.WebhookHeaderAttrType})
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
			WebhookName:     plan.Name.ValueString(),
			WebhookURL:      plan.URL.ValueString(),
			WebhookSecret:   plan.Secret.ValueString(),
			WebhookHeaders:  expandWebhookHeaders(ctx, plan.Headers, &resp.Diagnostics),
			WebhookPayloads: webhookTemplatesToClientPayloads(ctx, plan.Templates, plan.Variables, &resp.Diagnostics),
		},
	})
	if helper.AddDiagnosticOnError(&resp.Diagnostics, "Updating Honeycomb Webhook Recipient", err) {
		return
	}

	rcpt, err := r.client.Recipients.Get(ctx, plan.ID.ValueString())
	if helper.AddDiagnosticOnError(&resp.Diagnostics, "Updating Honeycomb Webhook Recipient", err) {
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

	// to prevent confusing if/else blocks, set null by default and override it if we have that detail on the recipient
	state.Templates = types.SetNull(types.ObjectType{AttrTypes: models.WebhookTemplateAttrType})
	state.Variables = types.SetNull(types.ObjectType{AttrTypes: models.TemplateVariableAttrType})
	state.Headers = types.SetNull(types.ObjectType{AttrTypes: models.WebhookHeaderAttrType})

	if rcpt.Details.WebhookPayloads != nil {
		state.Templates = plan.Templates
		if rcpt.Details.WebhookPayloads.TemplateVariables != nil {
			state.Variables = plan.Variables
		}
	}

	if rcpt.Details.WebhookHeaders != nil {
		state.Headers = plan.Headers
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

func webhookTemplatesToClientPayloads(ctx context.Context, templateSet types.Set, variableSet types.Set, diags *diag.Diagnostics) *client.WebhookPayloads {
	var templates []models.WebhookTemplateModel
	diags.Append(templateSet.ElementsAs(ctx, &templates, false)...)
	if diags.HasError() {
		return nil
	}

	var variables []models.TemplateVariableModel
	diags.Append(variableSet.ElementsAs(ctx, &variables, false)...)
	if diags.HasError() {
		return nil
	}

	clientWebhookPayloads := &client.WebhookPayloads{}

	for _, t := range templates {
		switch t.Type {
		case types.StringValue("trigger"):
			clientWebhookPayloads.PayloadTemplates.Trigger = &client.PayloadTemplate{
				Body: t.Body.ValueString(),
			}
		case types.StringValue("exhaustion_time"):
			clientWebhookPayloads.PayloadTemplates.ExhaustionTime = &client.PayloadTemplate{
				Body: t.Body.ValueString(),
			}
		case types.StringValue("budget_rate"):
			clientWebhookPayloads.PayloadTemplates.BudgetRate = &client.PayloadTemplate{
				Body: t.Body.ValueString(),
			}
		}
	}

	clientVars := make([]client.TemplateVariable, len(variables))
	for i, v := range variables {
		tmplVar := client.TemplateVariable{
			Name:    v.Name.ValueString(),
			Default: v.DefaultValue.ValueString(),
		}

		clientVars[i] = tmplVar
	}
	clientWebhookPayloads.TemplateVariables = clientVars

	return clientWebhookPayloads
}

func clientPayloadsToWebhookTemplateSets(ctx context.Context, p *client.WebhookPayloads, diags *diag.Diagnostics) (types.Set, types.Set) {
	if p == nil {
		return types.SetNull(types.ObjectType{AttrTypes: models.WebhookTemplateAttrType}), types.SetNull(types.ObjectType{AttrTypes: models.TemplateVariableAttrType})
	}

	tmplValues := webhookTemplatesToObjectValues(p.PayloadTemplates, diags)
	tmplResult, d := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: models.WebhookTemplateAttrType}, tmplValues)
	diags.Append(d...)

	var tmplVarValues []attr.Value
	for _, v := range p.TemplateVariables {
		tmplVarValues = append(tmplVarValues, webhookVariableToObjectValue(v, diags))
	}
	varResult, d := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: models.TemplateVariableAttrType}, tmplVarValues)
	diags.Append(d...)

	return tmplResult, varResult
}

func webhookTemplatesToObjectValues(templates client.PayloadTemplates, diags *diag.Diagnostics) []basetypes.ObjectValue {
	var templateObjs []basetypes.ObjectValue

	if templates.Trigger != nil {
		templateObjVal, d := types.ObjectValue(models.WebhookTemplateAttrType, map[string]attr.Value{
			"type": types.StringValue("trigger"),
			"body": types.StringValue(templates.Trigger.Body),
		})
		templateObjs = append(templateObjs, templateObjVal)
		diags.Append(d...)
	}

	if templates.BudgetRate != nil {
		templateObjVal, d := types.ObjectValue(models.WebhookTemplateAttrType, map[string]attr.Value{
			"type": types.StringValue("budget_rate"),
			"body": types.StringValue(templates.BudgetRate.Body),
		})
		templateObjs = append(templateObjs, templateObjVal)
		diags.Append(d...)
	}

	if templates.ExhaustionTime != nil {
		templateObjVal, d := types.ObjectValue(models.WebhookTemplateAttrType, map[string]attr.Value{
			"type": types.StringValue("exhaustion_time"),
			"body": types.StringValue(templates.ExhaustionTime.Body),
		})
		templateObjs = append(templateObjs, templateObjVal)
		diags.Append(d...)
	}

	return templateObjs
}

func webhookVariableToObjectValue(v client.TemplateVariable, diags *diag.Diagnostics) basetypes.ObjectValue {
	variableObj := map[string]attr.Value{
		"name":          types.StringValue(v.Name),
		"default_value": types.StringValue(v.Default),
	}
	varObjVal, d := types.ObjectValue(models.TemplateVariableAttrType, variableObj)
	diags.Append(d...)

	return varObjVal
}

func expandWebhookHeaders(ctx context.Context, set types.Set, diags *diag.Diagnostics) []client.WebhookHeader {
	var headers []models.WebhookHeaderModel
	diags.Append(set.ElementsAs(ctx, &headers, false)...)
	if diags.HasError() {
		return nil
	}

	clientHeaders := make([]client.WebhookHeader, len(headers))
	for i, h := range headers {
		hdr := client.WebhookHeader{
			Key:   h.Name.ValueString(),
			Value: h.Value.ValueString(),
		}

		clientHeaders[i] = hdr
	}

	return clientHeaders
}

func flattenWebhookHeaders(ctx context.Context, hdrs []client.WebhookHeader, diags *diag.Diagnostics) types.Set {
	var hdrValues []attr.Value
	for _, h := range hdrs {
		hdrValues = append(hdrValues, webhookHeaderToObjectValue(h, diags))
	}
	hdrResult, d := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: models.WebhookHeaderAttrType}, hdrValues)
	diags.Append(d...)

	return hdrResult
}

func webhookHeaderToObjectValue(h client.WebhookHeader, diags *diag.Diagnostics) basetypes.ObjectValue {
	headerObj := map[string]attr.Value{
		"name":  types.StringValue(h.Key),
		"value": types.StringValue(h.Value),
	}
	headerObjVal, d := types.ObjectValue(models.WebhookHeaderAttrType, headerObj)
	diags.Append(d...)

	return headerObjVal
}
