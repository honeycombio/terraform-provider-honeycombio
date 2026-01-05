package provider

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/validation"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/models"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                 = &flexibleBoardResource{}
	_ resource.ResourceWithConfigure    = &flexibleBoardResource{}
	_ resource.ResourceWithImportState  = &flexibleBoardResource{}
	_ resource.ResourceWithUpgradeState = &flexibleBoardResource{}
)

type flexibleBoardResource struct {
	client *client.Client
}

func NewFlexibleBoardResource() resource.Resource {
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
		Version:     1,
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
			"board_url": schema.StringAttribute{
				Computed:    true,
				Required:    false,
				Optional:    false,
				Description: "The URL of the Board in the Honeycomb UI.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"tags": tagsSchema(),
		},
		Blocks: map[string]schema.Block{
			"panel": schema.ListNestedBlock{
				Description: "List of panels to render on the board.",
				Validators: []validator.List{
					validation.RequireConsistentPanelPositions(),
					validation.RequirePanelTypeConsistency(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Required:    true,
							Description: `The panel type, either "query", "slo", or "text".`,
							Validators: []validator.String{
								stringvalidator.OneOf("query", "slo", "text"),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"position": schema.SingleNestedBlock{
							Description: `Manages the position of the panel on the board.`,
							Attributes: map[string]schema.Attribute{
								"x_coordinate": schema.Int64Attribute{
									Optional:    true,
									Computed:    true,
									Required:    false,
									Description: "The X coordinate of the panel.",
									Validators: []validator.Int64{
										int64validator.AtLeast(0),
										int64validator.AlsoRequires(path.MatchRelative().AtParent().AtName("y_coordinate")),
									},
								},
								"y_coordinate": schema.Int64Attribute{
									Optional:    true,
									Computed:    true,
									Required:    false,
									Description: "The Y coordinate of the panel.",
									Validators: []validator.Int64{
										int64validator.AtLeast(0),
										int64validator.AlsoRequires(path.MatchRelative().AtParent().AtName("x_coordinate")),
									},
								},
								"height": schema.Int64Attribute{
									Optional:    true,
									Computed:    true,
									Required:    false,
									Description: "The height of the panel.",
									Validators: []validator.Int64{
										int64validator.AtLeast(1),
									},
								},
								"width": schema.Int64Attribute{
									Optional:    true,
									Computed:    true,
									Required:    false,
									Description: "The width of the panel.",
									Validators: []validator.Int64{
										int64validator.AtLeast(1),
										int64validator.AtMost(12),
									},
								},
							},
						},
						"slo_panel": schema.ListNestedBlock{
							Description: "A Service Level Objective(SLO) panel to be displayed on the Board.",
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
										Optional:    false,
										Description: "Query annotation ID.",
									},
									"query_style": schema.StringAttribute{
										Optional:    true,
										Computed:    true,
										Description: "The visual style of the query (e.g., 'graph', 'combo').",
										Validators: []validator.String{
											stringvalidator.OneOf("graph", "table", "combo"),
										},
										Default: stringdefault.StaticString("graph"),
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
													Computed:    true,
													Description: "Display UTC Time X-Axis or Localtime X-Axis.",
												},
												"hide_markers": schema.BoolAttribute{
													Optional:    true,
													Computed:    true,
													Description: "Hide markers from appearing on graph.",
												},
												"hide_hovers": schema.BoolAttribute{
													Optional:    true,
													Computed:    true,
													Description: "Disable Graph tooltips in the results display when hovering over a graph.",
												},
												"prefer_overlaid_charts": schema.BoolAttribute{
													Optional:    true,
													Computed:    true,
													Description: "Combine any visualized AVG, MIN, MAX, and PERCENTILE clauses into a single chart.",
												},
												"hide_compare": schema.BoolAttribute{
													Optional:    true,
													Computed:    true,
													Description: "Hide comparison values.",
												},
											},
											Blocks: map[string]schema.Block{
												"chart": schema.ListNestedBlock{
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"chart_type": schema.StringAttribute{
																Optional:    true,
																Computed:    true,
																Description: "Type of chart (e.g., 'line', 'bar').",
																Validators: []validator.String{
																	stringvalidator.OneOf("default", "line", "tsbar", "stacked", "stat", "tsbar", "cpie", "cbar"),
																},
																Default: stringdefault.StaticString("default"),
															},
															"chart_index": schema.Int64Attribute{
																Optional:    true,
																Computed:    true,
																Description: "Index of the chart this configuration controls.",
																Validators: []validator.Int64{
																	int64validator.AtLeast(0),
																},
															},
															"omit_missing_values": schema.BoolAttribute{
																Optional:    true,
																Computed:    true,
																Description: "Omit missing values from the visualization.",
															},
															"use_log_scale": schema.BoolAttribute{
																Optional:    true,
																Computed:    true,
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
						"text_panel": schema.ListNestedBlock{
							Description: "A text panel to be displayed on the Board.",
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"content": schema.StringAttribute{
										Required:    true,
										Description: "The content of the text panel. Supports Markdown.",
									},
								},
							},
						},
					},
				},
			},
			"preset_filter": schema.ListNestedBlock{
				Description: "List of preset filters for the board. Maximum of 5 preset filters per board.",
				Validators: []validator.List{
					listvalidator.SizeAtMost(5),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"column": schema.StringAttribute{
							Required:    true,
							Description: "The column name for the preset filter.",
						},
						"alias": schema.StringAttribute{
							Required:    true,
							Description: "The alias for the preset filter.",
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

func (*flexibleBoardResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	// upgrade state as position is not an object and
	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema: &schema.Schema{
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
					"board_url": schema.StringAttribute{
						Computed:    true,
						Required:    false,
						Optional:    false,
						Description: "The URL of the Board in the Honeycomb UI.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"tags": tagsSchema(),
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
										listvalidator.SizeAtLeast(1),
									},
									NestedObject: schema.NestedBlockObject{
										Attributes: map[string]schema.Attribute{
											"x_coordinate": schema.Int64Attribute{
												Optional:    true,
												Required:    false,
												Computed:    true,
												Description: "The X coordinate of the panel.",
												Default:     int64default.StaticInt64(0),
												Validators: []validator.Int64{
													int64validator.AtLeast(0),
												},
											},
											"y_coordinate": schema.Int64Attribute{
												Optional:    true,
												Computed:    true,
												Required:    false,
												Description: "The Y coordinate of the panel.",
												Default:     int64default.StaticInt64(0),
												Validators: []validator.Int64{
													int64validator.AtLeast(0),
												},
											},
											"height": schema.Int64Attribute{
												Optional:    true,
												Computed:    true,
												Required:    false,
												Description: "The height of the panel.",
												Validators: []validator.Int64{
													int64validator.AtLeast(1),
												},
											},
											"width": schema.Int64Attribute{
												Optional:    true,
												Computed:    true,
												Required:    false,
												Description: "The width of the panel.",
												Validators: []validator.Int64{
													int64validator.AtLeast(1),
												},
											},
										},
									},
								},
								"slo_panel": schema.ListNestedBlock{
									Description: "A Service Level Objective(SLO) panel to be displayed on the Board.",
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
												Optional:    false,
												Description: "Query annotation ID.",
											},
											"query_style": schema.StringAttribute{
												Optional:    true,
												Computed:    true,
												Description: "The visual style of the query (e.g., 'graph', 'combo').",
												Validators: []validator.String{
													stringvalidator.OneOf("graph", "table", "combo"),
												},
												Default: stringdefault.StaticString("graph"),
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
															Computed:    true,
															Description: "Display UTC Time X-Axis or Localtime X-Axis.",
														},
														"hide_markers": schema.BoolAttribute{
															Optional:    true,
															Computed:    true,
															Description: "Hide markers from appearing on graph.",
														},
														"hide_hovers": schema.BoolAttribute{
															Optional:    true,
															Computed:    true,
															Description: "Disable Graph tooltips in the results display when hovering over a graph.",
														},
														"prefer_overlaid_charts": schema.BoolAttribute{
															Optional:    true,
															Computed:    true,
															Description: "Combine any visualized AVG, MIN, MAX, and PERCENTILE clauses into a single chart.",
														},
														"hide_compare": schema.BoolAttribute{
															Optional:    true,
															Computed:    true,
															Description: "Hide comparison values.",
														},
													},
													Blocks: map[string]schema.Block{
														"chart": schema.ListNestedBlock{
															NestedObject: schema.NestedBlockObject{
																Attributes: map[string]schema.Attribute{
																	"chart_type": schema.StringAttribute{
																		Optional:    true,
																		Computed:    true,
																		Description: "Type of chart (e.g., 'line', 'bar').",
																		Validators: []validator.String{
																			stringvalidator.OneOf("default", "line", "tsbar", "stacked", "stat", "tsbar", "cpie", "cbar"),
																		},
																		Default: stringdefault.StaticString("default"),
																	},
																	"chart_index": schema.Int64Attribute{
																		Optional:    true,
																		Computed:    true,
																		Description: "Index of the chart this configuration controls.",
																		Validators: []validator.Int64{
																			int64validator.AtLeast(0),
																		},
																	},
																	"omit_missing_values": schema.BoolAttribute{
																		Optional:    true,
																		Computed:    true,
																		Description: "Omit missing values from the visualization.",
																	},
																	"use_log_scale": schema.BoolAttribute{
																		Optional:    true,
																		Computed:    true,
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
			},
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				var oldState models.FlexibleBoardResourceModelV0
				resp.Diagnostics.Append(req.State.Get(ctx, &oldState)...)
				if resp.Diagnostics.HasError() {
					return
				}

				if oldState.Panels.IsNull() || oldState.Panels.IsUnknown() {
					return
				}

				// convert the old state to the new state
				var statePanels []models.FlexibleBoardPanelModelV0
				resp.Diagnostics.Append(oldState.Panels.ElementsAs(ctx, &statePanels, false)...)
				if resp.Diagnostics.HasError() {
					return
				}

				var newStatePanels []models.BoardPanelModel
				for _, panel := range statePanels {
					upgradedPanel := models.BoardPanelModel{
						PanelType:  panel.PanelType,
						QueryPanel: panel.QueryPanel,
						SLOPanel:   panel.SLOPanel,
						TextPanel:  types.ListNull(types.ObjectType{AttrTypes: models.TextPanelModelAttrType}),
					}

					if !panel.Position.IsNull() && !panel.Position.IsUnknown() {
						var oldStylePositions []models.BoardPanelPositionModel
						resp.Diagnostics.Append(panel.Position.ElementsAs(ctx, &oldStylePositions, false)...)
						if resp.Diagnostics.HasError() {
							return
						}

						if len(oldStylePositions) == 0 {
							upgradedPanel.Position = types.ObjectNull(models.BoardPanelPositionModelAttrType)
						} else {
							attrs := map[string]attr.Value{}
							attrs["x_coordinate"] = oldStylePositions[0].XCoordinate
							attrs["y_coordinate"] = oldStylePositions[0].YCoordinate
							attrs["height"] = oldStylePositions[0].Height
							attrs["width"] = oldStylePositions[0].Width

							obj, d := types.ObjectValue(models.BoardPanelPositionModelAttrType, attrs)
							resp.Diagnostics.Append(d...)

							upgradedPanel.Position = obj
						}
					} else {
						upgradedPanel.Position = types.ObjectNull(models.BoardPanelPositionModelAttrType)
					}

					newStatePanels = append(newStatePanels, upgradedPanel)
				}

				finalPanels, diags := types.ListValueFrom(ctx,
					types.ObjectType{AttrTypes: models.BoardPanelModelAttrType},
					newStatePanels,
				)
				resp.Diagnostics.Append(diags...)

				newState := models.FlexibleBoardResourceModel{
					ID:            oldState.ID,
					Name:          oldState.Name,
					Description:   oldState.Description,
					URL:           oldState.URL,
					Panels:        finalPanels,
					Tags:          oldState.Tags,
					PresetFilters: types.ListNull(types.ObjectType{AttrTypes: models.PresetFilterModelAttrType}),
				}
				resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
			},
		},
	}
}

func (r *flexibleBoardResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan, config models.FlexibleBoardResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	planTags, diags := helper.MapToTags(ctx, plan.Tags)
	if diags.HasError() {
		return
	}

	panelFromConfig := expandBoardPanels(ctx, plan.Panels, &resp.Diagnostics)
	finalPanels, layoutGeneration := setAPIDefaultsAndDetermineLayoutGeneration(panelFromConfig)
	presetFilters := expandPresetFilters(ctx, plan.PresetFilters, &resp.Diagnostics)
	createRequest := &client.Board{
		Name:             plan.Name.ValueString(),
		Description:      plan.Description.ValueString(),
		BoardType:        client.BoardTypeFlexible,
		Panels:           finalPanels,
		Tags:             planTags,
		LayoutGeneration: layoutGeneration,
		PresetFilters:    presetFilters,
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
	state.Description = types.StringValue(board.Description)
	state.URL = types.StringValue(board.Links.BoardURL)

	if len(board.Panels) == 0 {
		state.Panels = types.ListNull(types.ObjectType{AttrTypes: models.BoardPanelModelAttrType})
	} else {
		panelsObj := make([]attr.Value, 0, len(board.Panels))
		for i, panel := range board.Panels {
			panelValue := flattenBoardPanel(ctx, panel, &resp.Diagnostics, panelFromConfig[i])
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

	tags, diags := helper.TagsToMap(ctx, board.Tags)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Tags = tags

	state.PresetFilters = flattenPresetFilters(ctx, board.PresetFilters, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
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
	state.URL = types.StringValue(board.Links.BoardURL)

	if len(board.Panels) == 0 {
		state.Panels = types.ListNull(types.ObjectType{AttrTypes: models.BoardPanelModelAttrType})
	} else {
		var statePanels []models.BoardPanelModel
		resp.Diagnostics.Append(state.Panels.ElementsAs(ctx, &statePanels, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		panelConfig := expandBoardPanels(ctx, state.Panels, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}

		panelsObj := make([]attr.Value, 0, len(board.Panels))
		for i, panel := range board.Panels {
			var prevPosition client.BoardPanel
			if len(panelConfig) != 0 && i < len(panelConfig) {
				prevPosition = panelConfig[i]
			}
			panelValue := flattenBoardPanel(ctx, panel, &resp.Diagnostics, prevPosition)
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
	tags, diags := helper.TagsToMap(ctx, board.Tags)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Tags = tags

	state.PresetFilters = flattenPresetFilters(ctx, board.PresetFilters, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
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

	planTags, diags := helper.MapToTags(ctx, plan.Tags)
	if diags.HasError() {
		return
	}

	panelConfig := expandBoardPanels(ctx, plan.Panels, &resp.Diagnostics)
	finalPanels, layoutGeneration := setAPIDefaultsAndDetermineLayoutGeneration(panelConfig)
	presetFilters := expandPresetFilters(ctx, plan.PresetFilters, &resp.Diagnostics)
	updateRequest := &client.Board{
		ID:               plan.ID.ValueString(),
		Name:             plan.Name.ValueString(),
		Description:      plan.Description.ValueString(),
		BoardType:        client.BoardTypeFlexible,
		Panels:           finalPanels,
		Tags:             planTags,
		LayoutGeneration: layoutGeneration,
		PresetFilters:    presetFilters,
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
	state.URL = types.StringValue(board.Links.BoardURL)

	if len(board.Panels) == 0 {
		state.Panels = types.ListNull(types.ObjectType{AttrTypes: models.BoardPanelModelAttrType})
	} else {
		var statePanels []models.BoardPanelModel
		resp.Diagnostics.Append(config.Panels.ElementsAs(ctx, &statePanels, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		panelsObj := make([]attr.Value, 0, len(board.Panels))
		for i, panel := range board.Panels {
			panelValue := flattenBoardPanel(ctx, panel, &resp.Diagnostics, panelConfig[i])
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

	tags, diags := helper.TagsToMap(ctx, board.Tags)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Tags = tags

	state.PresetFilters = flattenPresetFilters(ctx, board.PresetFilters, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
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
			TextPanel:     expandBoardTextPanel(ctx, panel.TextPanel, diags),
		})
	}

	return result
}

// expandPanelPosition expands the panel position from the plan to the API model.
// It handles the case where the position is not set by setting X and Y to -1.
// This is a workaround for the limitations of the terraform v5 protocol.
func expandPanelPosition(
	ctx context.Context,
	panelPosition types.Object,
	diags *diag.Diagnostics,
) client.BoardPanelPosition {
	if panelPosition.IsNull() || panelPosition.IsUnknown() {
		return client.BoardPanelPosition{
			X:      -1,
			Y:      -1,
			Width:  -1,
			Height: -1,
		}
	}

	var position models.BoardPanelPositionModel
	diags.Append(panelPosition.As(ctx, &position, basetypes.ObjectAsOptions{})...)

	return client.BoardPanelPosition{
		X:      int(position.XCoordinate.ValueInt64()),
		Y:      int(position.YCoordinate.ValueInt64()),
		Height: int(position.Height.ValueInt64()),
		Width:  int(position.Width.ValueInt64()),
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

	return &client.BoardQueryPanel{
		QueryID:               queryPanels[0].QueryID.ValueString(),
		QueryAnnotationID:     queryPanels[0].QueryAnnotationID.ValueString(),
		Style:                 client.BoardQueryStyle(queryPanels[0].QueryStyle.ValueString()),
		VisualizationSettings: expandBoardQueryVisualizationSettings(ctx, queryPanels[0].VisualizationSettings, diags),
	}
}

func expandBoardQueryVisualizationSettings(
	ctx context.Context,
	settingsList types.List,
	diags *diag.Diagnostics,
) *client.BoardQueryVisualizationSettings {
	if settingsList.IsNull() || settingsList.IsUnknown() {
		return nil
	}

	var settings []models.VisualizationSettingsModel
	diags.Append(settingsList.ElementsAs(ctx, &settings, false)...)

	if len(settings) == 0 {
		return nil
	}

	return &client.BoardQueryVisualizationSettings{
		UseUTCXAxis:          settings[0].UseUTCXAxis.ValueBool(),
		HideMarkers:          settings[0].HideMarkers.ValueBool(),
		HideHovers:           settings[0].HideHovers.ValueBool(),
		PreferOverlaidCharts: settings[0].PreferOverlaidCharts.ValueBool(),
		HideCompare:          settings[0].HideCompare.ValueBool(),
		Charts:               expandBoardQueryVizCharts(ctx, settings[0].Charts, diags),
	}
}

func expandBoardQueryVizCharts(
	ctx context.Context,
	chartsList types.List,
	diags *diag.Diagnostics,
) []*client.ChartSettings {
	if chartsList.IsNull() || chartsList.IsUnknown() {
		return nil
	}
	var charts []models.ChartSettingsModel
	diags.Append(chartsList.ElementsAs(ctx, &charts, false)...)
	if len(charts) == 0 {
		return nil
	}

	result := make([]*client.ChartSettings, 0, len(charts))
	for _, chart := range charts {
		result = append(result, &client.ChartSettings{
			ChartType:         chart.ChartType.ValueString(),
			ChartIndex:        int(chart.ChartIndex.ValueInt64()),
			OmitMissingValues: chart.OmitMissingValues.ValueBool(),
			UseLogScale:       chart.LogScale.ValueBool(),
		})
	}

	return result
}

func expandBoardTextPanel(
	ctx context.Context,
	textPanel types.List,
	diags *diag.Diagnostics,
) *client.BoardTextPanel {

	if textPanel.IsNull() || textPanel.IsUnknown() {
		return nil
	}

	var textPanels []models.TextPanelModel
	diags.Append(textPanel.ElementsAs(ctx, &textPanels, false)...)

	if len(textPanels) == 0 {
		return nil
	}

	return &client.BoardTextPanel{
		Content: textPanels[0].Content.ValueString(),
	}
}

func expandPresetFilters(
	ctx context.Context,
	presetFilters types.List,
	diags *diag.Diagnostics,
) *[]client.PresetFilter {
	if presetFilters.IsNull() || presetFilters.IsUnknown() {
		emptySlice := []client.PresetFilter{}
		return &emptySlice
	}

	var filterModels []models.PresetFilterModel
	diags.Append(presetFilters.ElementsAs(ctx, &filterModels, false)...)
	if diags.HasError() {
		return nil
	}

	// If explicitly set to empty list, return empty slice to delete preset filters
	if len(filterModels) == 0 {
		emptySlice := []client.PresetFilter{}
		return &emptySlice
	}

	result := make([]client.PresetFilter, 0, len(filterModels))
	for _, filter := range filterModels {
		result = append(result, client.PresetFilter{
			Column: filter.Column.ValueString(),
			Alias:  filter.Alias.ValueString(),
		})
	}

	return &result
}

func flattenBoardPanel(
	ctx context.Context,
	panel client.BoardPanel,
	diags *diag.Diagnostics,
	statePanel client.BoardPanel,
) map[string]attr.Value {
	panelValue := make(map[string]attr.Value)
	panelValue["type"] = types.StringValue(string(panel.PanelType))
	panelValue["slo_panel"] = flattenBoardSloPanel(ctx, panel.SLOPanel, diags)
	panelValue["position"] = flattenBoardPanelPosition(panel.PanelPosition, diags, statePanel.PanelPosition)
	panelValue["query_panel"] = flattenBoardQueryPanel(ctx, panel.QueryPanel, diags)
	panelValue["text_panel"] = flattenBoardTextPanel(ctx, panel.TextPanel, diags)

	return panelValue
}

func flattenBoardPanelPosition(
	position client.BoardPanelPosition,
	diags *diag.Diagnostics,
	statePosition client.BoardPanelPosition,
) types.Object {
	// we use negative numbers to indicate that the panel position was never set. We use this to not write to state when panel position is not set.
	// This is a workaround for the various limitations that terraform v5 protocol presents.
	// Without this workaround, whenever the API generates a default position, terraform would complain about a schema mismatch between config and applied results.
	if statePosition.Height == -1 && statePosition.Width == -1 && statePosition.X == -1 && statePosition.Y == -1 {
		return types.ObjectNull(models.BoardPanelPositionModelAttrType)
	}
	x := position.X
	y := position.Y
	width := position.Width
	height := position.Height

	attrs := map[string]attr.Value{}
	if statePosition.X != -1 && statePosition.Y != -1 {
		attrs["x_coordinate"] = types.Int64Value(int64(x))
		attrs["y_coordinate"] = types.Int64Value(int64(y))
	}
	if statePosition.Width != -1 {
		attrs["width"] = types.Int64Value(int64(width))
	}
	if statePosition.Height != -1 {
		attrs["height"] = types.Int64Value(int64(height))
	}

	obj, d := types.ObjectValue(models.BoardPanelPositionModelAttrType, attrs)
	diags.Append(d...)

	return obj
}

func flattenBoardQueryPanel(
	ctx context.Context,
	queryPanel *client.BoardQueryPanel,
	diags *diag.Diagnostics,
) types.List {
	if queryPanel == nil {
		return types.ListNull(types.ObjectType{AttrTypes: models.QueryPanelModelAttrType})
	}

	obj, d := types.ObjectValue(models.QueryPanelModelAttrType, map[string]attr.Value{
		"query_id":               types.StringValue(queryPanel.QueryID),
		"query_annotation_id":    types.StringValue(queryPanel.QueryAnnotationID),
		"query_style":            types.StringValue(string(queryPanel.Style)),
		"visualization_settings": flattenBoardQueryVisualizationSettings(ctx, queryPanel.VisualizationSettings, diags),
	})
	diags.Append(d...)

	result, d := types.ListValueFrom(
		ctx,
		types.ObjectType{AttrTypes: models.QueryPanelModelAttrType},
		[]attr.Value{obj},
	)
	diags.Append(d...)

	return result
}

func flattenBoardQueryVisualizationSettings(
	ctx context.Context,
	settings *client.BoardQueryVisualizationSettings,
	diags *diag.Diagnostics,
) types.List {
	if settings == nil {
		return types.ListNull(types.ObjectType{AttrTypes: models.VisualizationSettingsModelAttrType})
	}

	obj, d := types.ObjectValue(models.VisualizationSettingsModelAttrType, map[string]attr.Value{
		"use_utc_xaxis":          types.BoolValue(settings.UseUTCXAxis),
		"hide_markers":           types.BoolValue(settings.HideMarkers),
		"hide_hovers":            types.BoolValue(settings.HideHovers),
		"prefer_overlaid_charts": types.BoolValue(settings.PreferOverlaidCharts),
		"hide_compare":           types.BoolValue(settings.HideCompare),
		"chart":                  flattenBoardQueryVizCharts(ctx, settings.Charts, diags),
	})
	diags.Append(d...)

	result, d := types.ListValueFrom(
		ctx,
		types.ObjectType{AttrTypes: models.VisualizationSettingsModelAttrType},
		[]attr.Value{obj},
	)
	diags.Append(d...)

	return result
}

func flattenBoardQueryVizCharts(
	ctx context.Context,
	charts []*client.ChartSettings,
	diags *diag.Diagnostics,
) types.List {
	if len(charts) == 0 {
		return types.ListNull(types.ObjectType{AttrTypes: models.ChartSettingsModelAttrType})
	}

	chartsObj := make([]attr.Value, 0, len(charts))
	for _, chart := range charts {
		if chart == nil {
			continue
		}

		obj, d := types.ObjectValue(models.ChartSettingsModelAttrType, map[string]attr.Value{
			"chart_type":          types.StringValue(chart.ChartType),
			"chart_index":         types.Int64Value(int64(chart.ChartIndex)),
			"omit_missing_values": types.BoolValue(chart.OmitMissingValues),
			"use_log_scale":       types.BoolValue(chart.UseLogScale),
		})
		diags.Append(d...)

		chartsObj = append(chartsObj, obj)
	}

	result, d := types.ListValueFrom(
		ctx,
		types.ObjectType{AttrTypes: models.ChartSettingsModelAttrType},
		chartsObj,
	)
	diags.Append(d...)

	return result
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

func flattenBoardTextPanel(
	ctx context.Context,
	textPanel *client.BoardTextPanel,
	diags *diag.Diagnostics,
) types.List {

	if textPanel == nil {
		return types.ListNull(types.ObjectType{AttrTypes: models.TextPanelModelAttrType})
	}

	obj, d := types.ObjectValue(models.TextPanelModelAttrType, map[string]attr.Value{
		"content": types.StringValue(textPanel.Content),
	})
	diags.Append(d...)

	result, d := types.ListValueFrom(
		ctx,
		types.ObjectType{AttrTypes: models.TextPanelModelAttrType},
		[]attr.Value{obj},
	)
	diags.Append(d...)

	return result
}

func flattenPresetFilters(
	ctx context.Context,
	presetFilters *[]client.PresetFilter,
	diags *diag.Diagnostics,
) types.List {
	if presetFilters == nil || len(*presetFilters) == 0 {
		return types.ListNull(types.ObjectType{AttrTypes: models.PresetFilterModelAttrType})
	}

	filtersObj := make([]attr.Value, 0, len(*presetFilters))
	for _, filter := range *presetFilters {
		obj, d := types.ObjectValue(models.PresetFilterModelAttrType, map[string]attr.Value{
			"column": types.StringValue(filter.Column),
			"alias":  types.StringValue(filter.Alias),
		})
		diags.Append(d...)

		filtersObj = append(filtersObj, obj)
	}

	result, d := types.ListValueFrom(
		ctx,
		types.ObjectType{AttrTypes: models.PresetFilterModelAttrType},
		filtersObj,
	)
	diags.Append(d...)

	return result
}

// setAPIDefaultsAndDetermineLayoutGeneration removes negative numbers. we use negative numbers to indicate that the panel position was never set.
// This is a workaround for the various limitations that terraform v5 protocol presents.
// The API will set default panel positions based on panel type. We decided not to write
// position to state when the panel position is not set.
// It returns layout generation "auto" if all panels have no positions.
func setAPIDefaultsAndDetermineLayoutGeneration(panels []client.BoardPanel) ([]client.BoardPanel, client.LayoutGeneration) {
	if len(panels) == 0 {
		return panels, client.LayoutGenerationAuto
	}
	resp := make([]client.BoardPanel, len(panels))
	copy(resp, panels)

	blankCount := 0

	for i := range resp {
		if resp[i].PanelPosition.X == -1 && resp[i].PanelPosition.Y == -1 {
			resp[i].PanelPosition.X = 0
			resp[i].PanelPosition.Y = 0
		}
		if resp[i].PanelPosition.Width == -1 {
			resp[i].PanelPosition.Width = 0
		}
		if resp[i].PanelPosition.Height == -1 {
			resp[i].PanelPosition.Height = 0
		}

		if resp[i].IsBlank() {
			blankCount++
		}
	}

	if blankCount == len(resp) {
		return resp, client.LayoutGenerationAuto
	}

	return resp, client.LayoutGenerationManual
}
