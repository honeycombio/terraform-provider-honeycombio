package provider

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/models"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &boardResource{}
	_ resource.ResourceWithConfigure   = &boardResource{}
	_ resource.ResourceWithImportState = &boardResource{}
)

type boardResource struct {
	client *client.Client
}

func NewBoardResource() resource.Resource {
	return &boardResource{}
}

func (*boardResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_board"
}

func (r *boardResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (*boardResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Board in a Honeycomb Environment.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Required:    false,
				Optional:    false,
				Description: "The ID of the Board.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the Board.",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The description of the Board. Supports Markdown.",
				Default:     stringdefault.StaticString(""),
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 1023),
				},
			},
			"column_layout": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The number of columns to layout on the Board.",
				Default:     stringdefault.StaticString(string(client.BoardColumnStyleSingle)),
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(client.BoardColumnStyleSingle),
						string(client.BoardColumnStyleMulti),
					),
				},
			},
			"style": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "How the Board should be displayed in the UI.",
				Default:     stringdefault.StaticString("visual"),
				DeprecationMessage: "All Boards are now displayed in the visual style. " +
					"Setting this value will have no effect. " +
					"This argument will be removed in a future version.",
				Validators: []validator.String{
					stringvalidator.OneOf(helper.AsStringSlice(client.BoardStyles())...),
				},
			},
			"board_url": schema.StringAttribute{
				Computed:    true,
				Required:    false,
				Optional:    false,
				Description: "The URL of the Board in the Honeycomb UI.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"slo": schema.SetNestedBlock{
				Description: "An SLO to be displayed on the Board.",
				Validators: []validator.Set{
					setvalidator.SizeAtMost(24),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Required:    true,
							Description: "The ID of the SLO.",
						},
					},
				},
			},
			"query": schema.ListNestedBlock{
				Description: "A query to be displayed on the Board.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"caption": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Default:     stringdefault.StaticString(""),
							Description: "Descriptive text to contextualize the Query within the Board. Supports Markdown.",
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 1023),
							},
						},
						"dataset": schema.StringAttribute{
							Optional:           true,
							Computed:           true,
							Description:        "The Dataset this Query is associated with.",
							DeprecationMessage: "Board Queries no longer require the dataset as they rely on the provided Query ID's dataset.",
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"query_id": schema.StringAttribute{
							Required:    true,
							Description: "The ID of the Query to run.",
						},
						"query_annotation_id": schema.StringAttribute{
							Optional:    true,
							Description: "The ID of the Query Annotation to associate with this Query.",
						},
						"query_style": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Default:     stringdefault.StaticString(string(client.BoardQueryStyleGraph)),
							Description: "How the query should be displayed within the Board.",
							Validators: []validator.String{
								stringvalidator.OneOf(helper.AsStringSlice(client.BoardQueryStyles())...),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"graph_settings": schema.ListNestedBlock{
							Description: `Manages the settings for this query's graph on the board.
See [Graph Settings](https://docs.honeycomb.io/working-with-your-data/graph-settings/) in the documentation for more information.`,
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							PlanModifiers: []planmodifier.List{
								//modifiers.DefaultGraphSettingsModifier(),
								listplanmodifier.UseStateForUnknown(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"log_scale": schema.BoolAttribute{
										Optional:    true,
										Computed:    true,
										Default:     booldefault.StaticBool(false),
										Description: "Set the graph's Y-axis to a logarithmic scale.",
									},
									"omit_missing_values": schema.BoolAttribute{
										Optional:    true,
										Computed:    true,
										Default:     booldefault.StaticBool(false),
										Description: "Enable interpolatation between datapoints when the intervening time buckets have no matching events.",
									},
									"hide_markers": schema.BoolAttribute{
										Optional:    true,
										Computed:    true,
										Default:     booldefault.StaticBool(false),
										Description: "Disable the overlay of Markers on the graph.",
									},
									"stacked_graphs": schema.BoolAttribute{
										Optional:    true,
										Computed:    true,
										Default:     booldefault.StaticBool(false),
										Description: "Enable the display of groups as stacked colored area under their line graphs.",
									},
									"utc_xaxis": schema.BoolAttribute{
										Optional:    true,
										Computed:    true,
										Default:     booldefault.StaticBool(false),
										Description: "Set the graph's X-axis to UTC.",
									},
									"overlaid_charts": schema.BoolAttribute{
										Optional:    true,
										Computed:    true,
										Default:     booldefault.StaticBool(false),
										Description: "Allow charts to be overlaid when using supported Visualize operators.",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *boardResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID == "" {
		resp.Diagnostics.AddError("Invalid Import ID", "The Board ID must be provided")
		return
	}

	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *boardResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan, config models.BoardResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createRequest := &client.Board{
		Name:         plan.Name.ValueString(),
		Description:  plan.Description.ValueString(),
		ColumnLayout: client.BoardColumnStyle(plan.ColumnLayout.ValueString()),
		Style:        client.BoardStyle(plan.Style.ValueString()),
		Queries:      expandBoardQueries(ctx, plan.Queries, &resp.Diagnostics),
		SLOs:         expandBoardSLOs(ctx, plan.SLOs, &resp.Diagnostics),
	}
	if resp.Diagnostics.HasError() {
		return
	}

	board, err := r.client.Boards.Create(ctx, createRequest)
	if helper.AddDiagnosticOnError(&resp.Diagnostics, "Creating Board", err) {
		return
	}

	var state models.BoardResourceModel
	state.ID = types.StringValue(board.ID)
	state.Name = types.StringValue(board.Name)
	state.Description = types.StringValue(board.Description)
	state.ColumnLayout = types.StringValue(string(board.ColumnLayout))
	state.Style = types.StringValue(string(board.Style))
	state.SLOs = flattenBoardSLOs(ctx, board.SLOs, &resp.Diagnostics)
	state.URL = types.StringValue(board.Links.BoardURL)

	if len(board.Queries) == 0 {
		state.Queries = types.ListNull(types.ObjectType{AttrTypes: models.BoardQueryModelAttrType})
	} else {
		var configuredQueries []models.BoardQueryModel
		resp.Diagnostics.Append(config.Queries.ElementsAs(ctx, &configuredQueries, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		querysObj := make([]attr.Value, 0, len(board.Queries))
		for i, query := range board.Queries {
			queryValue := flattenBoardQuery(ctx, query, &resp.Diagnostics)
			if resp.Diagnostics.HasError() {
				return
			}

			// GraphSettings are a special case as NestedBlockObjects cannot have a default value
			// and we need to differentiate between a Null value and a Zero value
			// to avoid inconsistent state issues
			if configuredQueries[i].GraphSettings.IsNull() {
				// if we've not configured a value, we'll keep the Null
				queryValue["graph_settings"] = configuredQueries[i].GraphSettings
			}

			obj, diag := types.ObjectValue(models.BoardQueryModelAttrType, queryValue)
			resp.Diagnostics.Append(diag...)

			querysObj = append(querysObj, obj)
		}

		queries, diag := types.ListValueFrom(ctx, types.ObjectType{
			AttrTypes: models.BoardQueryModelAttrType},
			querysObj,
		)
		resp.Diagnostics.Append(diag...)
		state.Queries = queries
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *boardResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.BoardResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var detailedErr client.DetailedError
	board, err := r.client.Boards.Get(ctx, state.ID.ValueString())
	if errors.As(err, &detailedErr) {
		if detailedErr.IsNotFound() {
			// if not found consider it delete -- so remove it from state
			resp.State.RemoveResource(ctx)
			return
		} else {
			resp.Diagnostics.Append(helper.NewDetailedErrorDiagnostic(
				"Error Reading Honeycomb Board",
				&detailedErr,
			))
		}
	} else if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Honeycomb Board",
			"Unexpected error reading Board ID "+state.ID.ValueString()+": "+err.Error(),
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	state.ID = types.StringValue(board.ID)
	state.Name = types.StringValue(board.Name)
	state.Description = types.StringValue(board.Description)
	state.ColumnLayout = types.StringValue(string(board.ColumnLayout))
	state.Style = types.StringValue(string(board.Style))
	state.SLOs = flattenBoardSLOs(ctx, board.SLOs, &resp.Diagnostics)
	state.URL = types.StringValue(board.Links.BoardURL)

	if len(board.Queries) == 0 {
		state.Queries = types.ListNull(types.ObjectType{AttrTypes: models.BoardQueryModelAttrType})
	} else {
		var stateQueries []models.BoardQueryModel
		resp.Diagnostics.Append(state.Queries.ElementsAs(ctx, &stateQueries, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		querysObj := make([]attr.Value, 0, len(board.Queries))
		for i, query := range board.Queries {
			queryValue := flattenBoardQuery(ctx, query, &resp.Diagnostics)
			if resp.Diagnostics.HasError() {
				return
			}

			// GraphSettings are a special case as NestedBlockObjects cannot have a default value
			// and we need to differentiate between a Null value and a Zero value
			// to avoid inconsistent state issues
			//
			// Additional length check is needed here to avoid index out of range
			// when doing an Import (which doesn't have state)
			if i < len(stateQueries) &&
				stateQueries[i].GraphSettings.IsNull() &&
				query.GraphSettings == (client.BoardGraphSettings{}) {
				// if we've written a Null value to the state, and
				// we've reading back a Zero value from the API: keep the Null
				queryValue["graph_settings"] = stateQueries[i].GraphSettings
			}

			obj, diag := types.ObjectValue(models.BoardQueryModelAttrType, queryValue)
			resp.Diagnostics.Append(diag...)

			querysObj = append(querysObj, obj)
		}

		queries, diag := types.ListValueFrom(ctx, types.ObjectType{
			AttrTypes: models.BoardQueryModelAttrType},
			querysObj,
		)
		resp.Diagnostics.Append(diag...)
		state.Queries = queries
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *boardResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, config models.BoardResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateRequest := &client.Board{
		ID:           plan.ID.ValueString(),
		Name:         plan.Name.ValueString(),
		Description:  plan.Description.ValueString(),
		ColumnLayout: client.BoardColumnStyle(plan.ColumnLayout.ValueString()),
		Style:        client.BoardStyle(plan.Style.ValueString()),
		Queries:      expandBoardQueries(ctx, plan.Queries, &resp.Diagnostics),
		SLOs:         expandBoardSLOs(ctx, plan.SLOs, &resp.Diagnostics),
	}
	if resp.Diagnostics.HasError() {
		return
	}

	board, err := r.client.Boards.Update(ctx, updateRequest)
	if helper.AddDiagnosticOnError(&resp.Diagnostics, "Updating Honeycomb Board", err) {
		return
	}

	var state models.BoardResourceModel
	state.ID = types.StringValue(board.ID)
	state.Name = types.StringValue(board.Name)
	state.Description = types.StringValue(board.Description)
	state.ColumnLayout = types.StringValue(string(board.ColumnLayout))
	state.Style = types.StringValue(string(board.Style))
	state.SLOs = flattenBoardSLOs(ctx, board.SLOs, &resp.Diagnostics)
	state.URL = types.StringValue(board.Links.BoardURL)

	if len(board.Queries) == 0 {
		state.Queries = types.ListNull(types.ObjectType{AttrTypes: models.BoardQueryModelAttrType})
	} else {
		var configuredQueries []models.BoardQueryModel
		resp.Diagnostics.Append(config.Queries.ElementsAs(ctx, &configuredQueries, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		querysObj := make([]attr.Value, 0, len(board.Queries))
		for i, query := range board.Queries {
			queryValue := flattenBoardQuery(ctx, query, &resp.Diagnostics)
			if resp.Diagnostics.HasError() {
				return
			}

			// GraphSettings are a special case as NestedBlockObjects cannot have a default value
			// and we need to differentiate between a Null value and a Zero value
			// to avoid inconsistent state issues
			if configuredQueries[i].GraphSettings.IsNull() {
				// if we've not configured a value, we'll keep the Null
				queryValue["graph_settings"] = configuredQueries[i].GraphSettings
			}

			obj, diag := types.ObjectValue(models.BoardQueryModelAttrType, queryValue)
			resp.Diagnostics.Append(diag...)

			querysObj = append(querysObj, obj)
		}

		queries, diag := types.ListValueFrom(
			ctx,
			types.ObjectType{AttrTypes: models.BoardQueryModelAttrType},
			querysObj,
		)
		resp.Diagnostics.Append(diag...)
		state.Queries = queries
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *boardResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state models.BoardResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var detailedErr client.DetailedError
	err := r.client.Boards.Delete(ctx, state.ID.ValueString())
	if err != nil {
		if errors.As(err, &detailedErr) {
			// if not found consider it deleted -- so don't error
			if !detailedErr.IsNotFound() {
				resp.Diagnostics.Append(helper.NewDetailedErrorDiagnostic(
					"Error Deleting Honeycomb Board",
					&detailedErr,
				))
			}
		} else {
			resp.Diagnostics.AddError(
				"Error Deleting Honeycomb Board",
				"Could not delete Board ID "+state.ID.ValueString()+": "+err.Error(),
			)
		}
	}
}

func expandBoardQueries(
	ctx context.Context,
	l types.List,
	diags *diag.Diagnostics,
) []client.BoardQuery {
	if l.IsNull() || l.IsUnknown() {
		return []client.BoardQuery{}
	}

	var queries []models.BoardQueryModel
	diags.Append(l.ElementsAs(ctx, &queries, false)...)
	if diags.HasError() {
		return nil
	}

	result := make([]client.BoardQuery, 0, len(queries))
	for _, query := range queries {
		var graphSettings []models.BoardQueryGraphSettingsModel
		diags.Append(query.GraphSettings.ElementsAs(ctx, &graphSettings, false)...)

		// if graph settings was empty, set to zero value
		if len(graphSettings) == 0 {
			graphSettings = []models.BoardQueryGraphSettingsModel{{}}
		}

		result = append(result, client.BoardQuery{
			QueryID:           query.ID.ValueString(),
			QueryAnnotationID: query.AnnotationID.ValueString(),
			Caption:           query.Caption.ValueString(),
			Dataset:           query.Dataset.ValueString(),
			QueryStyle:        client.BoardQueryStyle(query.Style.ValueString()),
			GraphSettings: client.BoardGraphSettings{
				UseLogScale:          graphSettings[0].LogScale.ValueBool(),
				OmitMissingValues:    graphSettings[0].OmitMissingValues.ValueBool(),
				HideMarkers:          graphSettings[0].HideMarkers.ValueBool(),
				UseStackedGraphs:     graphSettings[0].StackedGraphs.ValueBool(),
				UseUTCXAxis:          graphSettings[0].UTCXAxis.ValueBool(),
				PreferOverlaidCharts: graphSettings[0].OverlaidCharts.ValueBool(),
			},
		})
	}

	return result
}

func flattenBoardQuery(
	ctx context.Context,
	query client.BoardQuery,
	diags *diag.Diagnostics,
) map[string]attr.Value {
	queryValue := make(map[string]attr.Value)

	queryValue["caption"] = types.StringValue(query.Caption)
	queryValue["query_id"] = types.StringValue(query.QueryID)
	queryValue["query_style"] = types.StringValue(string(query.QueryStyle))
	queryValue["dataset"] = types.StringValue(query.Dataset)
	if query.QueryAnnotationID == "" {
		queryValue["query_annotation_id"] = types.StringNull()
	} else {
		queryValue["query_annotation_id"] = types.StringValue(query.QueryAnnotationID)
	}

	graphObj, diag := types.ObjectValue(models.BoardQueryGraphSettingsModelAttrType, map[string]attr.Value{
		"log_scale":           types.BoolValue(query.GraphSettings.UseLogScale),
		"omit_missing_values": types.BoolValue(query.GraphSettings.OmitMissingValues),
		"hide_markers":        types.BoolValue(query.GraphSettings.HideMarkers),
		"stacked_graphs":      types.BoolValue(query.GraphSettings.UseStackedGraphs),
		"utc_xaxis":           types.BoolValue(query.GraphSettings.UseUTCXAxis),
		"overlaid_charts":     types.BoolValue(query.GraphSettings.PreferOverlaidCharts),
	})
	diags.Append(diag...)

	graphSettings, diag := types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: models.BoardQueryGraphSettingsModelAttrType},
		[]attr.Value{graphObj},
	)
	diags.Append(diag...)

	queryValue["graph_settings"] = graphSettings

	return queryValue
}

func expandBoardSLOs(
	ctx context.Context,
	s types.Set,
	diags *diag.Diagnostics,
) []string {
	if s.IsNull() || s.IsUnknown() {
		return []string{}
	}

	var slos []models.BoardSLOModel
	diags.Append(s.ElementsAs(ctx, &slos, false)...)
	if diags.HasError() {
		return nil
	}

	result := make([]string, 0, len(slos))
	for _, slo := range slos {
		result = append(result, slo.ID.ValueString())
	}

	return result
}

func flattenBoardSLOs(
	ctx context.Context,
	slos []string,
	diags *diag.Diagnostics,
) types.Set {
	if len(slos) == 0 {
		return types.SetNull(types.ObjectType{AttrTypes: models.BoardSLOModelAttrType})
	}

	slosObj := make([]attr.Value, 0, len(slos))
	for _, slo := range slos {
		obj, d := types.ObjectValue(models.BoardSLOModelAttrType, map[string]attr.Value{
			"id": types.StringValue(slo),
		})
		diags.Append(d...)

		if d.HasError() {
			continue
		}
		slosObj = append(slosObj, obj)
	}

	result, d := types.SetValueFrom(
		ctx,
		types.ObjectType{AttrTypes: models.BoardSLOModelAttrType},
		slosObj,
	)
	diags.Append(d...)

	return result
}
