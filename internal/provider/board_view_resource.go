package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/coerce"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/models"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &boardViewResource{}
	_ resource.ResourceWithConfigure   = &boardViewResource{}
	_ resource.ResourceWithImportState = &boardViewResource{}
	_ resource.ResourceWithValidateConfig = &boardViewResource{}
)

type boardViewResource struct {
	client *client.Client
}

func NewBoardViewResource() resource.Resource {
	return &boardViewResource{}
}

func (*boardViewResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_board_view"
}

func (r *boardViewResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (*boardViewResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a board view in a Honeycomb flexible board.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the board view.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"board_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the flexible board this view belongs to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the board view.",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"filter": schema.ListNestedBlock{
				Description:         "List of filters to apply to the board view. **Required:** At least one filter must be specified.",
				MarkdownDescription: "List of filters to apply to the board view. **Required:** At least one filter must be specified.",
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"column": schema.StringAttribute{
							Description: "The column to filter on.",
							Required:    true,
						},
						"operation": schema.StringAttribute{
							Description:         "The operator to apply.",
							MarkdownDescription: "The operator to apply. See the supported list at [Filter Operators](https://docs.honeycomb.io/api/query-specification/#filter-operators). Not all operators require a value.",
							Required:            true,
							Validators: []validator.String{
								stringvalidator.OneOf(helper.AsStringSlice(client.FilterOps())...),
							},
						},
						"value": schema.StringAttribute{ // TODO: convert to DynamicAttribute when supported in nested blocks
							Description: "The value used for the filter. Not needed if operation is \"exists\" or \"does-not-exist\". For \"in\" or \"not-in\" operations, provide a comma-separated list of values.",
							Optional:    true,
						},
					},
				},
			},
		},
	}
}

func (r *boardViewResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config models.BoardViewResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate filters if present
	if !config.Filters.IsNull() && !config.Filters.IsUnknown() {
		var filterModels []models.BoardViewFilterModel
		resp.Diagnostics.Append(config.Filters.ElementsAs(ctx, &filterModels, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		for i, f := range filterModels {
			filterOp := client.FilterOpFromString(f.Operation.ValueString())
			if filterOp == client.FilterOp("") {
				// Skip invalid operations - they'll be caught by schema validation
				continue
			}

			// Validate array operations for empty strings in comma-separated values
			if filterOp.IsArray() && !f.Value.IsNull() && !f.Value.IsUnknown() {
				values := strings.Split(f.Value.ValueString(), ",")
				for j, value := range values {
					trimmed := strings.TrimSpace(value)
					if trimmed == "" {
						resp.Diagnostics.AddAttributeError(
							path.Root("filter").AtListIndex(i).AtName("value"),
							"Empty value in comma-separated list",
							fmt.Sprintf("operation '%s' does not allow empty values in the comma-separated list (found empty value at position %d)", f.Operation.ValueString(), j+1),
						)
					}
				}
			}
		}
	}
}

func (r *boardViewResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: board_id/view_id
	boardID, viewID, found := strings.Cut(req.ID, "/")
	if !found {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Board view import ID must be in the format: board_id/view_id",
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &models.BoardViewResourceModel{
		ID:      types.StringValue(viewID),
		BoardID: types.StringValue(boardID),
		Name:    types.StringUnknown(), // Will be populated by Read
		Filters: types.ListNull(types.ObjectType{AttrTypes: models.BoardViewFilterModelAttrType}),
	})...)
}

func (r *boardViewResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan models.BoardViewResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that at least one filter is provided
	if plan.Filters.IsNull() || plan.Filters.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("filter"),
			"At least one filter is required",
			"Board views must have at least one filter defined.",
		)
		return
	}

	var filterModels []models.BoardViewFilterModel
	resp.Diagnostics.Append(plan.Filters.ElementsAs(ctx, &filterModels, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(filterModels) == 0 {
		resp.Diagnostics.AddAttributeError(
			path.Root("filter"),
			"At least one filter is required",
			"Board views must have at least one filter defined.",
		)
		return
	}

	filters := expandBoardViewFilters(ctx, plan.Filters, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	createRequest := &client.BoardView{
		Name:    plan.Name.ValueString(),
		Filters: filters,
	}

	boardView, err := r.client.BoardViews.Create(ctx, plan.BoardID.ValueString(), createRequest)
	if helper.AddDiagnosticOnError(&resp.Diagnostics, "Creating Board View", err) {
		return
	}

	state := flattenBoardView(ctx, boardView, plan.BoardID, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *boardViewResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.BoardViewResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var detailedErr client.DetailedError
	boardView, err := r.client.BoardViews.Get(ctx, state.BoardID.ValueString(), state.ID.ValueString())
	if errors.As(err, &detailedErr) {
		if detailedErr.IsNotFound() {
			// if not found consider it deleted -- remove it from state
			resp.State.RemoveResource(ctx)
			return
		} else {
			resp.Diagnostics.Append(helper.NewDetailedErrorDiagnostic(
				"Error Reading Honeycomb Board View",
				&detailedErr,
			))
		}
	} else if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Honeycomb Board View",
			"Unexpected error reading Board View ID "+state.ID.ValueString()+": "+err.Error(),
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	newState := flattenBoardView(ctx, boardView, state.BoardID, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
}

func (r *boardViewResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan models.BoardViewResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that at least one filter is provided
	if plan.Filters.IsNull() || plan.Filters.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("filter"),
			"At least one filter is required",
			"Board views must have at least one filter defined.",
		)
		return
	}

	var filterModels []models.BoardViewFilterModel
	resp.Diagnostics.Append(plan.Filters.ElementsAs(ctx, &filterModels, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(filterModels) == 0 {
		resp.Diagnostics.AddAttributeError(
			path.Root("filter"),
			"At least one filter is required",
			"Board views must have at least one filter defined.",
		)
		return
	}

	filters := expandBoardViewFilters(ctx, plan.Filters, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	updateRequest := &client.BoardView{
		ID:      plan.ID.ValueString(),
		Name:    plan.Name.ValueString(),
		Filters: filters,
	}

	boardView, err := r.client.BoardViews.Update(ctx, plan.BoardID.ValueString(), updateRequest)
	if helper.AddDiagnosticOnError(&resp.Diagnostics, "Updating Board View", err) {
		return
	}

	state := flattenBoardView(ctx, boardView, plan.BoardID, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *boardViewResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state models.BoardViewResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var detailedErr client.DetailedError
	err := r.client.BoardViews.Delete(ctx, state.BoardID.ValueString(), state.ID.ValueString())
	if err != nil {
		if errors.As(err, &detailedErr) {
			// if not found consider it deleted -- so don't error
			if !detailedErr.IsNotFound() {
				resp.Diagnostics.Append(helper.NewDetailedErrorDiagnostic(
					"Error Deleting Honeycomb Board View",
					&detailedErr,
				))
			}
		} else {
			resp.Diagnostics.AddError(
				"Error Deleting Honeycomb Board View",
				"Could not delete Board View ID "+state.ID.ValueString()+": "+err.Error(),
			)
		}
	}
}

// expandBoardViewFilters converts Terraform filter models to API filter format
func expandBoardViewFilters(
	ctx context.Context,
	filtersList types.List,
	diags *diag.Diagnostics,
) []client.BoardViewFilter {
	if filtersList.IsNull() || filtersList.IsUnknown() {
		return []client.BoardViewFilter{}
	}

	var filterModels []models.BoardViewFilterModel
	diags.Append(filtersList.ElementsAs(ctx, &filterModels, false)...)
	if diags.HasError() {
		return nil
	}

	filters := make([]client.BoardViewFilter, 0, len(filterModels))
	for i, f := range filterModels {
		filterOp := client.FilterOpFromString(f.Operation.ValueString())
		if filterOp == client.FilterOp("") {
			diags.AddAttributeError(
				path.Root("filter").AtListIndex(i).AtName("operation"),
				"Invalid filter operation",
				fmt.Sprintf("'%s' is not a valid filter operation", f.Operation.ValueString()),
			)
			continue
		}

		filter := client.BoardViewFilter{
			Column:    f.Column.ValueString(),
			Operation: f.Operation.ValueString(),
		}

		// Handle value conversion based on operation type
		if !f.Value.IsNull() && !f.Value.IsUnknown() {
			if filterOp.IsArray() {
				// For array operations, expect comma-separated string
				values := strings.Split(f.Value.ValueString(), ",")
				// Validate that there are no empty strings (after trimming)
				hasEmpty := false
				result := make([]any, 0, len(values))
				for j, value := range values {
					trimmed := strings.TrimSpace(value)
					if trimmed == "" {
						hasEmpty = true
						diags.AddAttributeError(
							path.Root("filter").AtListIndex(i).AtName("value"),
							"Empty value in comma-separated list",
							fmt.Sprintf("operation '%s' does not allow empty values in the comma-separated list (found empty value at position %d)", f.Operation.ValueString(), j+1),
						)
					} else {
						result = append(result, coerce.ValueToType(trimmed))
					}
				}
				// Validate that at least one non-empty value was provided
				if len(result) == 0 {
					diags.AddAttributeError(
						path.Root("filter").AtListIndex(i).AtName("value"),
						"Empty filter value",
						fmt.Sprintf("operation '%s' requires at least one non-empty value in the comma-separated list", f.Operation.ValueString()),
					)
				} else if !hasEmpty {
					// Only set filter.Value if validation passed (no empty strings)
					filter.Value = result
				}
			} else {
				// For scalar operations, convert string to appropriate type
				filter.Value = coerce.ValueToType(f.Value.ValueString())
			}
		}

		// Validate filter value based on operation type
		if filterOp.IsUnary() {
			if filter.Value != nil {
				diags.AddAttributeError(
					path.Root("filter").AtListIndex(i).AtName("value"),
					f.Operation.ValueString()+" does not take a value",
					"",
				)
			}
		} else {
			if filter.Value == nil {
				diags.AddAttributeError(
					path.Root("filter").AtListIndex(i).AtName("operation"),
					"operator "+f.Operation.ValueString()+" requires a value",
					"",
				)
			}
		}

		filters = append(filters, filter)
	}

	return filters
}

// flattenBoardView converts API board view to Terraform model
func flattenBoardView(
	ctx context.Context,
	boardView *client.BoardView,
	boardID types.String,
	diags *diag.Diagnostics,
) models.BoardViewResourceModel {
	state := models.BoardViewResourceModel{
		ID:      types.StringValue(boardView.ID),
		BoardID: boardID,
		Name:    types.StringValue(boardView.Name),
	}

	if len(boardView.Filters) == 0 {
		state.Filters = types.ListNull(types.ObjectType{AttrTypes: models.BoardViewFilterModelAttrType})
	} else {
		filtersObj := make([]attr.Value, 0, len(boardView.Filters))
		for _, filter := range boardView.Filters {
			// Convert filter value to string
			var valueStr types.String
			if filter.Value == nil {
				valueStr = types.StringNull()
			} else {
				// Convert the value to string representation
				switch v := filter.Value.(type) {
				case []any:
					// Array value - convert to comma-separated string
					strValues := make([]string, len(v))
					for i, elem := range v {
						strValues[i] = formatFilterValueToString(elem)
					}
					valueStr = types.StringValue(strings.Join(strValues, ","))
				default:
					// Convert to string with proper formatting
					valueStr = types.StringValue(formatFilterValueToString(filter.Value))
				}
			}

			obj, d := types.ObjectValue(models.BoardViewFilterModelAttrType, map[string]attr.Value{
				"column":    types.StringValue(filter.Column),
				"operation": types.StringValue(filter.Operation),
				"value":     valueStr,
			})
			diags.Append(d...)

			filtersObj = append(filtersObj, obj)
		}

		filters, d := types.ListValueFrom(
			ctx,
			types.ObjectType{AttrTypes: models.BoardViewFilterModelAttrType},
			filtersObj,
		)
		diags.Append(d...)
		state.Filters = filters
	}

	return state
}

// formatFilterValueToString converts a filter value to string, formatting numbers
// without unnecessary decimal places to avoid round-trip issues.
func formatFilterValueToString(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case int, int32, int64:
		return fmt.Sprintf("%d", v)
	case float32:
		// Format float32 without unnecessary decimals
		if float64(v) == float64(int64(v)) {
			return fmt.Sprintf("%d", int64(v))
		}
		return fmt.Sprintf("%g", v)
	case float64:
		// Format float64 without unnecessary decimals
		if v == float64(int64(v)) {
			return fmt.Sprintf("%d", int64(v))
		}
		return fmt.Sprintf("%g", v)
	case bool:
		return fmt.Sprintf("%t", v)
	default:
		// Fallback to coerce helper for other types
		return coerce.ValueToString(value)
	}
}
