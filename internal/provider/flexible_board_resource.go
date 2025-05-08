package provider

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
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
	_ resource.Resource                = &flexibleBoardResource{}
	_ resource.ResourceWithConfigure   = &flexibleBoardResource{}
	_ resource.ResourceWithImportState = &flexibleBoardResource{}
)

type flexibleBoardResource struct {
	client *client.Client
}

func NewflexibleBoardResource() resource.Resource {
	return &flexibleBoardResource{}
}

func (*flexibleBoardResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_flexible_board"
}

func (r *flexibleBoardResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (*flexibleBoardResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a flexible Board in a Honeycomb Environment.",
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
			"type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The type of board.",
				Default:     stringdefault.StaticString("flexible"),
				Validators: []validator.String{
					stringvalidator.OneOf("flexible"),
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
			"panel": schema.ListNestedBlock{
				Description: "List of panels to render on the board.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Required:    true,
							Description: `The panel type, either "query" or "slo".`,
							Validators: []validator.String{
								stringvalidator.OneOf("query", "slo"),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"position": schema.ListNestedBlock{
							Description: `Manages the position of the panel on the board.`,
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"x_coordinate": schema.Int64Attribute{
										Optional:    true,
										Required:    false,
										Computed:    true,
										Description: "The X coordinate of the panel.",
									},
									"y_coordinate": schema.Int64Attribute{
										Optional:    true,
										Computed:    true,
										Description: "The Y coordinate of the panel.",
									},
									"height": schema.Int64Attribute{
										Optional:    true,
										Computed:    true,
										Required:    false,
										Description: "The height of the panel.",
									},
									"width": schema.Int64Attribute{
										Optional:    true,
										Computed:    true,
										Description: "The width of the panel.",
									},
								},
							},
						},
						"slo_panel": schema.ListNestedBlock{
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"slo_id": schema.StringAttribute{
										Required:    true,
										Description: "SLO ID to display in this panel.",
									},
								},
							},
						},
						"query_panel": schema.ListNestedBlock{
							Description: "A query panel to be displayed on the Board.",
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"query_id": schema.StringAttribute{
										Required:    true,
										Description: "Query ID to be rendered in the panel.",
									},
									"query_annotation_id": schema.StringAttribute{
										Required:    true,
										Description: "Query annotation ID.",
									},
									"dataset": schema.StringAttribute{
										Optional:    true,
										Description: "Dataset associated with the query.",
									},
									"query_style": schema.StringAttribute{
										Required:    true,
										Description: "The visual style of the query (e.g., 'graph', 'combo').",
									},
								},
								Blocks: map[string]schema.Block{
									"visualization_settings": schema.ListNestedBlock{
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"use_utc_xaxis": schema.BoolAttribute{
													Optional:    true,
													Description: "Render the X axis in UTC time.",
												},
												"hide_markers": schema.BoolAttribute{
													Optional:    true,
													Description: "Hide markers on the chart.",
												},
												"hide_hovers": schema.BoolAttribute{
													Optional:    true,
													Description: "Disable hover tooltips.",
												},
												"prefer_overlaid_charts": schema.BoolAttribute{
													Optional:    true,
													Description: "Prefer overlaid rendering for multiple charts.",
												},
												"hide_compare": schema.BoolAttribute{
													Optional:    true,
													Description: "Hide comparison values.",
												},
											},
											Blocks: map[string]schema.Block{
												"chart": schema.ListNestedBlock{
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"chart_type": schema.StringAttribute{
																Optional:    true,
																Description: "Type of chart (e.g., 'line', 'bar').",
															},
															"chart_index": schema.Int64Attribute{
																Required:    true,
																Description: "Index of the chart in the layout.",
															},
															"omit_missing_values": schema.BoolAttribute{
																Optional:    true,
																Computed:    true,
																Default:     booldefault.StaticBool(false),
																Description: "Omit missing values from the visualization.",
															},
															"use_log_scale": schema.BoolAttribute{
																Optional:    true,
																Computed:    true,
																Default:     booldefault.StaticBool(false),
																Description: "Use logarithmic scale on Y axis.",
															},
														},
													},
												},
											},
										},
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

func (r *flexibleBoardResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID == "" {
		resp.Diagnostics.AddError("Invalid Import ID", "The Board ID must be provided")
		return
	}

	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *flexibleBoardResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan, config models.FlexibleBoardResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createRequest := &client.Board{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		BoardType:   plan.BoardType.ValueString(),
		Panels:      expandBoardPanels(ctx, plan.Panels, &resp.Diagnostics),
	}

	if resp.Diagnostics.HasError() {
		return
	}

	board, err := r.client.Boards.Create(ctx, createRequest)
	if helper.AddDiagnosticOnError(&resp.Diagnostics, "Creating Board", err) {
		return
	}

	var state models.FlexibleBoardResourceModel
	state.ID = types.StringValue(board.ID)
	state.Name = types.StringValue(board.Name)
	state.BoardType = types.StringValue(board.BoardType)
	state.Description = types.StringValue(board.Description)
	state.URL = types.StringValue(board.Links.BoardURL)

	if len(board.Panels) == 0 {
		state.Panels = types.ListNull(types.ObjectType{AttrTypes: models.BoardPanelModelAttrType})
	} else {
		panelsObj := make([]attr.Value, 0, len(board.Panels))
		for _, panel := range board.Panels {
			panelValue := flattenBoardPanel(ctx, panel, &resp.Diagnostics)
			if resp.Diagnostics.HasError() {
				return
			}

			obj, diag := types.ObjectValue(models.BoardPanelModelAttrType, panelValue)
			resp.Diagnostics.Append(diag...)

			panelsObj = append(panelsObj, obj)
		}

		panels, diag := types.ListValueFrom(ctx,
			types.ObjectType{AttrTypes: models.BoardPanelModelAttrType},
			panelsObj,
		)
		resp.Diagnostics.Append(diag...)
		state.Panels = panels
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *flexibleBoardResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.FlexibleBoardResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var detailedErr client.DetailedError
	board, err := r.client.Boards.Get(ctx, state.ID.ValueString())
	if errors.As(err, &detailedErr) {
		if detailedErr.IsNotFound() {
			// if not found consider it deleted -- remove it from state
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
	state.BoardType = types.StringValue(board.BoardType)
	state.URL = types.StringValue(board.Links.BoardURL)

	if len(board.Panels) == 0 {
		state.Panels = types.ListNull(types.ObjectType{AttrTypes: models.BoardPanelModelAttrType})
	} else {
		var statePanels []models.BoardPanelModel
		resp.Diagnostics.Append(state.Panels.ElementsAs(ctx, &statePanels, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		panelsObj := make([]attr.Value, 0, len(board.Panels))
		for _, panel := range board.Panels {
			panelValue := flattenBoardPanel(ctx, panel, &resp.Diagnostics)
			if resp.Diagnostics.HasError() {
				return
			}

			obj, diag := types.ObjectValue(models.BoardPanelModelAttrType, panelValue)
			resp.Diagnostics.Append(diag...)

			panelsObj = append(panelsObj, obj)
		}

		panels, diag := types.ListValueFrom(ctx,
			types.ObjectType{AttrTypes: models.BoardPanelModelAttrType},
			panelsObj,
		)
		resp.Diagnostics.Append(diag...)
		state.Panels = panels
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *flexibleBoardResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, config models.FlexibleBoardResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateRequest := &client.Board{
		ID:          plan.ID.ValueString(),
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		BoardType:   plan.BoardType.ValueString(),
		Panels:      expandBoardPanels(ctx, plan.Panels, &resp.Diagnostics),
	}
	if resp.Diagnostics.HasError() {
		return
	}

	board, err := r.client.Boards.Update(ctx, updateRequest)
	if helper.AddDiagnosticOnError(&resp.Diagnostics, "Updating Honeycomb Board", err) {
		return
	}

	var state models.FlexibleBoardResourceModel
	state.ID = types.StringValue(board.ID)
	state.Name = types.StringValue(board.Name)
	state.Description = types.StringValue(board.Description)
	state.BoardType = types.StringValue(board.BoardType)
	state.URL = types.StringValue(board.Links.BoardURL)

	if len(board.Panels) == 0 {
		state.Panels = types.ListNull(types.ObjectType{AttrTypes: models.BoardPanelModelAttrType})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *flexibleBoardResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state models.FlexibleBoardResourceModel
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

func expandBoardPanels(
	ctx context.Context,
	l types.List,
	diags *diag.Diagnostics,
) []client.BoardPanel {
	if l.IsNull() || l.IsUnknown() {
		return []client.BoardPanel{}
	}

	var panels []models.BoardPanelModel
	diags.Append(l.ElementsAs(ctx, &panels, false)...)
	if diags.HasError() {
		return nil
	}

	result := make([]client.BoardPanel, 0, len(panels))
	for _, panel := range panels {
		result = append(result, client.BoardPanel{
			PanelType:     client.BoardPanelType(panel.PanelType.ValueString()),
			PanelPosition: expandPanelPosition(ctx, panel.Position, diags),
			SLOPanel:      expandBoardSLOPanel(ctx, panel.SLOPanel, diags),
			QueryPanel:    expandBoardQueryPanel(ctx, panel.QueryPanel, diags),
		})
	}

	return result
}

func expandPanelPosition(
	ctx context.Context,
	panelPosition types.List,
	diags *diag.Diagnostics,
) client.BoardPanelPosition {
	if panelPosition.IsNull() || panelPosition.IsUnknown() {
		return client.BoardPanelPosition{}
	}

	var position []models.BoardPanelPositionModel
	diags.Append(panelPosition.ElementsAs(ctx, &position, false)...)

	if len(position) == 0 {
		position = []models.BoardPanelPositionModel{{}}
	}

	return client.BoardPanelPosition{
		X:      int(position[0].XCoordinate.ValueInt64()),
		Y:      int(position[0].YCoordinate.ValueInt64()),
		Height: int(position[0].Height.ValueInt64()),
		Width:  int(position[0].Width.ValueInt64()),
	}
}

func expandBoardSLOPanel(
	ctx context.Context,
	sloPanel types.List,
	diags *diag.Diagnostics,
) *client.BoardSLOPanel {

	if sloPanel.IsNull() || sloPanel.IsUnknown() {
		return nil
	}

	var sloPanels []models.SLOPanelModel
	diags.Append(sloPanel.ElementsAs(ctx, &sloPanels, false)...)

	if len(sloPanels) == 0 {
		return nil
	}
	return &client.BoardSLOPanel{
		SLOID: sloPanels[0].SLOID.ValueString(),
	}
}

func expandBoardQueryPanel(
	ctx context.Context,
	queryPanel types.List,
	diags *diag.Diagnostics,
) *client.BoardQueryPanel {
	if queryPanel.IsNull() || queryPanel.IsUnknown() {
		return nil
	}

	var queryPanels []models.QueryPanelModel
	diags.Append(queryPanel.ElementsAs(ctx, &queryPanels, false)...)

	if len(queryPanels) == 0 {
		return nil
	}
	return &client.BoardQueryPanel{}
}

func flattenBoardPanel(
	ctx context.Context,
	panel client.BoardPanel,
	diags *diag.Diagnostics,
) map[string]attr.Value {
	panelValue := make(map[string]attr.Value)
	panelValue["type"] = types.StringValue(string(panel.PanelType))
	panelValue["slo_panel"] = flattenBoardSloPanel(ctx, panel.SLOPanel, diags)
	panelValue["position"] = flattenBoardPanelPosition(ctx, panel.PanelPosition, diags)
	panelValue["query_panel"] = flattenBoardQueryPanel(ctx, panel.QueryPanel, diags)

	return panelValue
}

func flattenBoardPanelPosition(
	ctx context.Context,
	position client.BoardPanelPosition,
	diags *diag.Diagnostics,
) types.List {
	obj, d := types.ObjectValue(models.BoardPanelPositionModelAttrType, map[string]attr.Value{
		"x_coordinate": types.Int64Value(int64(position.X)),
		"y_coordinate": types.Int64Value(int64(position.Y)),
		"height":       types.Int64Value(int64(position.Height)),
		"width":        types.Int64Value(int64(position.Width)),
	})
	diags.Append(d...)

	result, d := types.ListValueFrom(
		ctx,
		types.ObjectType{AttrTypes: models.BoardPanelPositionModelAttrType},
		[]attr.Value{obj},
	)
	diags.Append(d...)

	return result
}

func flattenBoardQueryPanel(
	_ context.Context,
	_ *client.BoardQueryPanel,
	_ *diag.Diagnostics,
) types.List {
	// if queryPanel == nil {

	// }

	return types.ListNull(types.ObjectType{AttrTypes: models.QueryPanelModelAttrType})

	// obj, d := types.ObjectValue(models.QueryPanelModelAttrType, map[string]attr.Value{
	// 	"query_id":         types.StringValue(queryPanel.QueryID),
	// 	"query_annotation_id": types.StringValue(queryPanel.QueryAnnotationID),
	// 	"dataset":          types.StringValue(queryPanel.Dataset),
	// 	"query_style":      types.StringValue(string(queryPanel.Style)),
	// })
	// diags.Append(d...)

	// result, d := types.ListValueFrom(
	// 	ctx,
	// 	types.ObjectType{AttrTypes: models.QueryPanelModelAttrType},
	// 	[]attr.Value{obj},
	// )
	// diags.Append(d...)

	// return result
}

func flattenBoardSloPanel(
	ctx context.Context,
	sloPanel *client.BoardSLOPanel,
	diags *diag.Diagnostics,
) types.List {
	if sloPanel == nil {
		return types.ListNull(types.ObjectType{AttrTypes: models.SLOPanelModelAttrType})
	}

	obj, d := types.ObjectValue(models.SLOPanelModelAttrType, map[string]attr.Value{
		"slo_id": types.StringValue(sloPanel.SLOID),
	})
	diags.Append(d...)

	result, d := types.ListValueFrom(
		ctx,
		types.ObjectType{AttrTypes: models.SLOPanelModelAttrType},
		[]attr.Value{obj},
	)
	diags.Append(d...)

	return result
}
