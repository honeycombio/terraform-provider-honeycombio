package provider

import (
	"context"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/coerce"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/hashcode"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/validation"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/models"
)

// Ensure the implementation satisfies the expected interfaces.
var _ datasource.DataSource = &querySpecDataSource{}

// relationalFieldPrefixes are the prefixes that indicate a field references a related span.
var relationalFieldPrefixes = []string{"root.", "child.", "parent.", "any.", "any2.", "any3.", "none."}

// isRelationalField returns true if the column name references a related span.
func isRelationalField(column string) bool {
	for _, prefix := range relationalFieldPrefixes {
		if strings.HasPrefix(column, prefix) {
			return true
		}
	}
	return false
}

func NewQuerySpecDataSource() datasource.DataSource {
	return &querySpecDataSource{}
}

type querySpecDataSource struct{}

func (d *querySpecDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_query_specification"
}

func (d *querySpecDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Generates a Query Specification in JSON format for use with resources that expect a JSON-formatted Query Specification like \"honeycombio_query\". " +
			"Using this data source to generate query specifications is optional. " +
			"It is also valid to use literal JSON strings in your configuration or to use the \"file\" interpolation function to read a raw JSON query specification from a file.",
		MarkdownDescription: "Generates a [Query Specification](https://docs.honeycomb.io/api/query-specification/) in JSON format for use with resources that expect a " +
			"JSON-formatted Query Specification like [`honeycombio_query`](../resources/query.md). " +
			"Using this data source to generate query specifications is optional. " +
			"It is also valid to use literal JSON strings in your configuration or to use the `file` interpolation function to read a raw JSON query specification from a file.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:        "The ID of the query specification.",
				DeprecationMessage: "The `id` attribute is deprecated and included for compatibility with the Terraform Plugin SDK. It will be removed in a future version.",
				Computed:           true,
				Required:           false,
				Optional:           false,
			},
			"filter_combination": schema.StringAttribute{
				Description: "How to combine multiple filters. Defaults to \"AND\".",
				Optional:    true,
				Validators: []validator.String{stringvalidator.OneOf(
					string(client.FilterCombinationAnd), string(client.FilterCombinationOr),
				)},
			},
			"breakdowns": schema.ListAttribute{
				Description: "A list of fields to group results by.",
				ElementType: types.StringType,
				Optional:    true,
			},
			"limit": schema.Int64Attribute{
				Description: "The maximum number of results to return. Defaults to 1000.",
				Optional:    true,
				Validators:  []validator.Int64{int64validator.Between(1, client.DefaultQueryLimit)},
			},
			"time_range": schema.Int64Attribute{
				Description: "The time range of the query, in seconds. Defaults to 7200.",
				Optional:    true,
			},
			"start_time": schema.Int64Attribute{
				Description: "The absolute start time of the query's time range, in seconds since the Unix epoch.",
				Optional:    true,
			},
			"end_time": schema.Int64Attribute{
				Description: "The absolute end time of the query's time range, in seconds since the Unix epoch.",
				Optional:    true,
			},
			"granularity": schema.Int64Attribute{
				Description: "The time resolution of the query's graph, in seconds. " +
					"Valid values must be in between the query's time range at maximum, and /1000 at minimum.",
				Optional:   true,
				Validators: []validator.Int64{int64validator.AtLeast(0)},
			},
			"compare_time_offset": schema.Int64Attribute{
				Description: "The time offset for comparison queries, in seconds. " +
					"Used to compare current time range data with data from a previous time period.",
				Optional: true,
			},
			"json": schema.StringAttribute{
				Description:         "The generated query specification in JSON format.",
				MarkdownDescription: "JSON representation of the query according to the [Query Specification](https://docs.honeycomb.io/api/query-specification/#fields-on-a-query-specification), can be used as input for other resources.",
				Computed:            true,
				Required:            false,
				Optional:            false,
			},
		},
		Blocks: map[string]schema.Block{
			"calculation": schema.ListNestedBlock{
				Description: "Zero or more configuration blocks describing the calculations to return as a time series and summary table. " +
					"If no calculations are provided, \"COUNT\" is assumed.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"op": schema.StringAttribute{
							Description:         "The operator to apply.",
							MarkdownDescription: "The operator to apply. See the supported list at [Calculation Operators](https://docs.honeycomb.io/api/query-specification/#calculation-operators).",
							Required:            true,
							Validators: []validator.String{
								stringvalidator.OneOf(helper.AsStringSlice(client.CalculationOps())...),
							},
						},
						"column": schema.StringAttribute{
							Description: "The column to apply the operator on. " +
								"Not allowed with \"COUNT\" or \"CONCURRENCY\", required for all other operators.",
							Optional: true,
						},
						"name": schema.StringAttribute{
							Description: "The name of the calculation. Required when using calculation filters or when referencing the calculation in a formula.",
							Optional:    true,
						},
						"filter_combination": schema.StringAttribute{
							Description: "How to combine multiple calculation filters. Defaults to \"AND\".",
							Optional:    true,
							Validators: []validator.String{stringvalidator.OneOf(
								helper.AsStringSlice(client.FilterCombinations())...,
							)},
						},
					},
					Blocks: map[string]schema.Block{
						"filter": schema.ListNestedBlock{
							Description: "Zero or more configuration blocks describing filters to apply to this specific calculation.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"column": schema.StringAttribute{
										Description: "The column to filter on.",
										Required:    true,
									},
									"op": schema.StringAttribute{
										Description:         "The operator to apply.",
										MarkdownDescription: "The operator to apply. See the supported list at [Filter Operators](https://docs.honeycomb.io/api/query-specification/#filter-operators). Not all operators require a value.",
										Required:            true,
										Validators: []validator.String{
											stringvalidator.OneOf(helper.AsStringSlice(client.FilterOps())...),
										},
									},
									"value": schema.StringAttribute{
										Description: "The value used for the filter. Not needed if op is \"exists\" or \"does-not-exist\".",
										Optional:    true,
									},
								},
							},
						},
					},
				},
			},
			"calculated_field": schema.ListNestedBlock{
				Description: "Zero or more configuration blocks describing the Calculated Fields to create for use in this query.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "The name of the Calculated Field.",
							Required:    true,
						},
						"expression": schema.StringAttribute{
							Description:         "The formula to use for the Calculated Field.",
							MarkdownDescription: "The formula to use for the Calculated Field. See the [Calculated Field Formula Reference](https://docs.honeycomb.io/reference/derived-column-formula/) for syntax and functions.",
							Required:            true,
							Validators: []validator.String{
								validation.IsValidCalculatedField(),
							},
						},
					},
				},
			},
			"formula": schema.ListNestedBlock{
				Description: "Zero or more configuration blocks describing formulas that compute values from calculations.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "The name of the formula.",
							Required:    true,
						},
						"expression": schema.StringAttribute{
							Description:         "An expression that references calculations by name using the calculated field syntax.",
							MarkdownDescription: "An expression that references calculations by name. Uses the same syntax as [Calculated Field expressions](https://docs.honeycomb.io/reference/derived-column-formula/), but references calculation names instead of column names.",
							Required:            true,
						},
					},
				},
			},
			"filter": schema.ListNestedBlock{
				Description: "Zero or more configuration blocks describing the filters to apply to the query results.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"column": schema.StringAttribute{
							Description: "The column to filter on.",
							Required:    true,
						},
						"op": schema.StringAttribute{
							Description:         "The operator to apply.",
							MarkdownDescription: "The operator to apply. See the supported list at [Filter Operators](https://docs.honeycomb.io/api/query-specification/#filter-operators). Not all operators require a value.",
							Required:            true,
							Validators: []validator.String{
								stringvalidator.OneOf(helper.AsStringSlice(client.FilterOps())...),
							},
						},
						"value": schema.StringAttribute{ // TODO: convert to DynamicAttribute
							Description: "The value used for the filter. Not needed if op is \"exists\" or \"not-exists\".",
							Optional:    true,
						},
					},
				},
			},
			"having": schema.ListNestedBlock{
				Description: "Zero or more configuration blocks used to restrict returned groups in the query result.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"calculate_op": schema.StringAttribute{
							Description:         "The operator to apply.",
							MarkdownDescription: "The calculation operator to apply. Supports all [Calculation Operators](https://docs.honeycomb.io/api/query-specification/#calculation-operators) except `HEATMAP`.",
							Required:            true,
							Validators: []validator.String{
								stringvalidator.OneOf(helper.AsStringSlice(client.HavingCalculationOps())...),
							},
						},
						"column": schema.StringAttribute{
							Description: "The column to filter on. Not allowed with \"COUNT\".",
							Optional:    true,
						},
						"op": schema.StringAttribute{
							Description: "The operator to apply.",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOf(helper.AsStringSlice(client.HavingOps())...),
							},
						},
						"value": schema.Float64Attribute{ // The API currently assumes this is a number
							Description: "The value used for the filter.",
							Required:    true,
						},
					},
				},
			},
			"order": schema.ListNestedBlock{
				Description: "Zero or more configuration blocks describing how to order the query results. " +
					"Each term must appear as a \"calculation\" or in \"breakdowns\".",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"column": schema.StringAttribute{
							Description: "Either a column present in \"breakdown\" or a column that \"op\" applies to.",
							Optional:    true,
						},
						"op": schema.StringAttribute{
							Description:         "The operator to apply.",
							MarkdownDescription: "The calculation operator to apply. See the supported list at [Calculation Operators](https://docs.honeycomb.io/api/query-specification/#calculation-operators).",
							Optional:            true,
							Validators: []validator.String{
								stringvalidator.OneOf(helper.AsStringSlice(client.CalculationOps())...),
							},
						},
						"order": schema.StringAttribute{
							Description: "The sort order to apply.",
							Optional:    true,
							Validators: []validator.String{
								stringvalidator.OneOf(helper.AsStringSlice(client.SortOrders())...),
							},
						},
					},
				},
			},
		},
	}
}

func (d *querySpecDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data models.QuerySpecificationModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	calculations := make([]client.CalculationSpec, 0, len(data.Calculations))
	for i, c := range data.Calculations {
		calculation := client.CalculationSpec{
			Op:     client.CalculationOp(c.Op.ValueString()),
			Column: c.Column.ValueStringPointer(),
			Name:   c.Name.ValueStringPointer(),
		}

		if calculation.Op.IsUnaryOp() && calculation.Column != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("calculation").AtListIndex(i).AtName("column"),
				"column is not allowed with operator "+c.Op.ValueString(),
				"",
			)
		} else if !calculation.Op.IsUnaryOp() && calculation.Column == nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("calculation").AtListIndex(i).AtName("op"),
				c.Op.ValueString()+" requires a column",
				"",
			)
		}

		// Process calculation filters
		if len(c.Filters) > 0 {
			// Require name when using filters
			if calculation.Name == nil {
				resp.Diagnostics.AddAttributeError(
					path.Root("calculation").AtListIndex(i).AtName("name"),
					"name is required when using calculation filters",
					"",
				)
			}

			calcFilters := make([]client.FilterSpec, 0, len(c.Filters))
			for j, f := range c.Filters {
				filter := client.FilterSpec{
					Column: f.Column.ValueString(),
					Op:     client.FilterOp(f.Op.ValueString()),
				}

				// Validate that relational fields are not used in calculation filters
				if isRelationalField(filter.Column) {
					resp.Diagnostics.AddAttributeError(
						path.Root("calculation").AtListIndex(i).AtName("filter").AtListIndex(j).AtName("column"),
						"relational fields are not supported in calculation filters",
						"columns prefixed with 'root.', 'child.', 'parent.', etc cannot be used in calculation filters",
					)
				}

				// Process filter value
				if !f.Value.IsNull() {
					if filter.Op == client.FilterOpIn || filter.Op == client.FilterOpNotIn {
						values := strings.Split(f.Value.ValueString(), ",")
						result := make([]any, len(values))
						for k, value := range values {
							result[k] = coerce.ValueToType(value)
						}
						filter.Value = result
					} else {
						filter.Value = coerce.ValueToType(f.Value.ValueString())
					}
				}

				// Validate filter op/value combinations
				if filter.Op == client.FilterOpExists || filter.Op == client.FilterOpDoesNotExist {
					if filter.Value != nil {
						resp.Diagnostics.AddAttributeError(
							path.Root("calculation").AtListIndex(i).AtName("filter").AtListIndex(j).AtName("value"),
							f.Op.ValueString()+" does not take a value",
							"",
						)
					}
				} else {
					if filter.Value == nil {
						resp.Diagnostics.AddAttributeError(
							path.Root("calculation").AtListIndex(i).AtName("filter").AtListIndex(j).AtName("op"),
							"operator "+f.Op.ValueString()+" requires a value",
							"",
						)
					}
				}

				calcFilters = append(calcFilters, filter)
			}
			calculation.Filters = calcFilters
			calculation.FilterCombination = client.FilterCombination(c.FilterCombination.ValueString())
		}

		calculations = append(calculations, calculation)
	}
	// 'COUNT' is the default calculation and will be returned by the API if
	// none have been provided. As this can potentially cause an infinite diff
	// we'll set the default here if we haven't parsed any
	if len(calculations) == 0 {
		calculations = []client.CalculationSpec{{Op: client.CalculationOpCount}}
	}

	// Track all names used by calculations and formulas (must be unique across both)
	type nameSource struct {
		sourceType string // "calculation" or "formula"
		index      int
	}
	namesSeen := make(map[string]nameSource)

	// Check for duplicate calculation names
	for i, calc := range calculations {
		if calc.Name != nil && *calc.Name != "" {
			if prev, exists := namesSeen[*calc.Name]; exists {
				resp.Diagnostics.AddAttributeError(
					path.Root("calculation").AtListIndex(i).AtName("name"),
					"duplicate name",
					"name \""+*calc.Name+"\" is already used by "+prev.sourceType+" at index "+strconv.Itoa(prev.index),
				)
			} else {
				namesSeen[*calc.Name] = nameSource{sourceType: "calculation", index: i}
			}
		}
	}

	calculatedFields := make([]client.CalculatedFieldSpec, len(data.CalculatedFields))
	for i, f := range data.CalculatedFields {
		calculatedFields[i] = client.CalculatedFieldSpec{
			Name:       f.Name.ValueString(),
			Expression: f.Expression.ValueString(),
		}
	}

	formulas := make([]client.FormulaSpec, len(data.Formulas))
	for i, f := range data.Formulas {
		formulas[i] = client.FormulaSpec{
			Name:       f.Name.ValueString(),
			Expression: f.Expression.ValueString(),
		}
	}

	// Check for duplicate formula names (including conflicts with calculation names)
	for i, formula := range formulas {
		if formula.Name != "" {
			if prev, exists := namesSeen[formula.Name]; exists {
				resp.Diagnostics.AddAttributeError(
					path.Root("formula").AtListIndex(i).AtName("name"),
					"duplicate name",
					"name \""+formula.Name+"\" is already used by "+prev.sourceType+" at index "+strconv.Itoa(prev.index),
				)
			} else {
				namesSeen[formula.Name] = nameSource{sourceType: "formula", index: i}
			}
		}
	}

	// Validate that formulas cannot be used with HEATMAP calculations
	if len(formulas) > 0 {
		for i, calc := range calculations {
			if calc.Op == client.CalculationOpHeatmap {
				resp.Diagnostics.AddAttributeError(
					path.Root("calculation").AtListIndex(i).AtName("op"),
					"HEATMAP calculations cannot be used with formulas",
					"formulas are not supported when any calculation uses the HEATMAP operator",
				)
			}
		}
	}

	// Check if any calculation has filters (needed for relational field validation later)
	hasCalculationFilters := false
	for _, calc := range calculations {
		if len(calc.Filters) > 0 {
			hasCalculationFilters = true
			break
		}
	}

	filters := make([]client.FilterSpec, 0, len(data.Filters))
	for i, f := range data.Filters {
		filter := client.FilterSpec{
			Column: f.Column.ValueString(),
			Op:     client.FilterOp(f.Op.ValueString()),
		}

		// Validate that relational fields are not used in filters when formulas or calculation filters are present
		if (len(formulas) > 0 || hasCalculationFilters) && isRelationalField(filter.Column) {
			resp.Diagnostics.AddAttributeError(
				path.Root("filter").AtListIndex(i).AtName("column"),
				"relational fields are not supported when using formulas or calculation filters",
				"columns prefixed with 'root.', 'child.', 'parent.', etc cannot be used in filters when formulas or calculation filters are present",
			)
		}

		// TODO: replace with DynamicAttribute
		if !f.Value.IsNull() {
			if filter.Op == client.FilterOpIn || filter.Op == client.FilterOpNotIn {
				// if the filter operation is 'in' or 'not-in' we expect the value
				// to be a CSV string so we build it into a slice
				values := strings.Split(f.Value.ValueString(), ",")
				result := make([]any, len(values))
				for i, value := range values {
					result[i] = coerce.ValueToType(value)
				}
				filter.Value = result
			} else {
				filter.Value = coerce.ValueToType(f.Value.ValueString())
			}
		}

		if filter.Op == client.FilterOpExists || filter.Op == client.FilterOpDoesNotExist {
			if filter.Value != nil {
				resp.Diagnostics.AddAttributeError(
					path.Root("filter").AtListIndex(i).AtName("value"),
					f.Op.ValueString()+" does not take a value",
					"",
				)
			}
		} else {
			if filter.Value == nil {
				resp.Diagnostics.AddAttributeError(
					path.Root("filter").AtListIndex(i).AtName("op"),
					"operator "+f.Op.ValueString()+" requires a value",
					"",
				)
			}
		}

		filters = append(filters, filter)
	}

	havings := make([]client.HavingSpec, 0, len(data.Havings))
	for i, h := range data.Havings {
		having := client.HavingSpec{
			CalculateOp: client.ToPtr(client.CalculationOp(h.CalculateOp.ValueString())),
			Column:      h.Column.ValueStringPointer(),
			Op:          client.ToPtr(client.HavingOp(h.Op.ValueString())),
			Value:       h.Value.ValueFloat64(),
		}

		if having.CalculateOp.IsUnaryOp() && having.Column != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("having").AtListIndex(i).AtName("calculate_op"),
				h.CalculateOp.ValueString()+" should not have an accompanying column",
				"",
			)
		}
		if !having.CalculateOp.IsUnaryOp() && having.Column == nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("having").AtListIndex(i).AtName("calculate_op"),
				h.CalculateOp.ValueString()+" requires a column",
				"",
			)
		}

		havings = append(havings, having)
	}
	// Ensure all havings have a matching calculate_op/column pair
	for i, having := range havings {
		found := false
		for _, calc := range calculations {
			if reflect.DeepEqual(having.Column, calc.Column) &&
				*having.CalculateOp == calc.Op {
				found = true
				break
			}
		}
		if !found {
			resp.Diagnostics.AddAttributeError(
				path.Root("having").AtListIndex(i).AtName("calculate_op"),
				string(*having.CalculateOp)+" missing matching calculation",
				"each having must have a matching calculation",
			)
		}
	}

	breakdowns := make([]string, len(data.Breakdowns))
	for i, b := range data.Breakdowns {
		breakdowns[i] = b.ValueString()

		// Validate that relational fields are not used in breakdowns when formulas or calculation filters are present
		if (len(formulas) > 0 || hasCalculationFilters) && isRelationalField(breakdowns[i]) {
			resp.Diagnostics.AddAttributeError(
				path.Root("breakdowns").AtListIndex(i),
				"relational fields are not supported when using formulas or calculation filters",
				"columns prefixed with 'root.', 'child.', 'parent.', etc cannot be used in breakdowns when formulas or calculation filters are present",
			)
		}
	}

	orders := make([]client.OrderSpec, len(data.Orders))
	for i, o := range data.Orders {
		order := client.OrderSpec{
			Column: o.Column.ValueStringPointer(),
		}
		if !o.Op.IsNull() {
			order.Op = client.ToPtr(client.CalculationOp(o.Op.ValueString()))
		}

		// ascending is the default, API doesn't return or require
		// the field unless value is descending
		//
		// not sending to avoid constant plan diffs
		if !o.Order.IsNull() {
			ov := client.SortOrder(o.Order.ValueString())
			if ov != client.SortOrderAsc {
				order.Order = &ov
			}
		}

		orders[i] = order
	}
	// ensure all orders have a matching calculation, calculated_field, formula, or breakdown
	for i, order := range orders {
		if order.Op != nil && *order.Op == client.CalculationOpHeatmap {
			resp.Diagnostics.AddAttributeError(
				path.Root("order").AtListIndex(i).AtName("op"),
				"cannot order by HEATMAP",
				"",
			)
			continue
		}
		found := false
		for _, calc := range calculations {
			if reflect.DeepEqual(order.Column, calc.Column) {
				found = true
				break
			}
			// Also check if order references a named calculation
			if order.Column != nil && calc.Name != nil && *order.Column == *calc.Name {
				found = true
				break
			}
		}
		if !found {
			for _, calc := range calculatedFields {
				if order.Column != nil && *order.Column == calc.Name {
					found = true
					break
				}
			}
		}
		if !found {
			for _, formula := range formulas {
				if order.Column != nil && *order.Column == formula.Name {
					found = true
					break
				}
			}
		}
		if !found {
			for _, breakdown := range breakdowns {
				if order.Column != nil && *order.Column == breakdown {
					found = true
					break
				}
			}
		}
		if !found {
			resp.Diagnostics.AddAttributeError(
				path.Root("order").AtListIndex(i),
				"missing matching calculation, formula, or breakdown",
				"each order must have a matching calculation, formula, or breakdown",
			)
		}
	}

	querySpec := &client.QuerySpec{
		Calculations:      calculations,
		CalculatedFields:  calculatedFields,
		Formulas:          formulas,
		Filters:           filters,
		Havings:           havings,
		FilterCombination: client.FilterCombination(data.FilterCombination.ValueString()),
		Breakdowns:        breakdowns,
		Orders:            orders,
		StartTime:         data.StartTime.ValueInt64Pointer(),
		EndTime:           data.EndTime.ValueInt64Pointer(),
	}
	if !data.Limit.IsNull() {
		querySpec.Limit = client.ToPtr(int(data.Limit.ValueInt64()))
	}
	if data.TimeRange.IsNull() {
		querySpec.TimeRange = client.ToPtr(client.DefaultQueryTimeRange)
	} else {
		querySpec.TimeRange = client.ToPtr(int(data.TimeRange.ValueInt64()))
	}
	if !data.Granularity.IsNull() {
		querySpec.Granularity = client.ToPtr(int(data.Granularity.ValueInt64()))
	}
	if !data.CompareTimeOffset.IsNull() {
		timeOffset := data.CompareTimeOffset.ValueInt64()

		// get the query time range to check if the compare time offset is greater
		var queryTimeRange int64
		if !data.TimeRange.IsUnknown() && !data.TimeRange.IsNull() {
			queryTimeRange = data.TimeRange.ValueInt64()
		} else if !data.StartTime.IsUnknown() && !data.StartTime.IsNull() && !data.EndTime.IsUnknown() && !data.EndTime.IsNull() {
			queryTimeRange = data.EndTime.ValueInt64() - data.StartTime.ValueInt64()
		} else {
			queryTimeRange = client.DefaultQueryTimeRange
		}

		if timeOffset < queryTimeRange {
			resp.Diagnostics.AddAttributeError(
				path.Root("compare_time_offset"),
				"compare_time_offset must be greater than the queries time range.",
				"",
			)
		}
		if !slices.Contains(client.ValidTimeCompareOffsets, data.CompareTimeOffset.ValueInt64()) {
			validTimeOffsets := append([]int64{queryTimeRange}, client.ValidTimeCompareOffsets...)
			// Convert int to string for the error message
			timeOverTimeStrings := make([]string, len(validTimeOffsets))
			for i, val := range validTimeOffsets {
				timeOverTimeStrings[i] = strconv.FormatInt(val, 10)
			}

			resp.Diagnostics.AddAttributeError(
				path.Root("compare_time_offset"),
				"compare_time_offset is an invalid value.",
				"Valid values are: "+strings.Join(timeOverTimeStrings, ", "),
			)
		}

		querySpec.CompareTimeOffsetSeconds = client.ToPtr(int(data.CompareTimeOffset.ValueInt64()))
	}

	if querySpec.TimeRange != nil && querySpec.StartTime != nil && querySpec.EndTime != nil {
		resp.Diagnostics.AddError(
			"invalid time configuration",
			"specify at most two of time_range, start_time and end_time",
		)
	}
	if querySpec.TimeRange != nil && querySpec.Granularity != nil {
		if *querySpec.Granularity > *querySpec.TimeRange {
			resp.Diagnostics.AddAttributeError(
				path.Root("granularity"),
				"invalid granularity",
				"granularity can not be greater than time_range",
			)
		}
		if *querySpec.Granularity != 0 && *querySpec.Granularity < (*querySpec.TimeRange/1000) {
			resp.Diagnostics.AddAttributeError(
				path.Root("granularity"),
				"invalid granularity",
				"granularity can not be less than time_range/1000",
			)
		}
	}

	// if we encountered any errors during parsing, we'll stop here
	if resp.Diagnostics.HasError() {
		return
	}

	json, err := querySpec.Encode()
	if err != nil {
		resp.Diagnostics.AddError(
			"Encoding query specification",
			err.Error(),
		)
		return
	}
	data.Json = types.StringValue(json)
	data.ID = types.StringValue(strconv.Itoa(hashcode.String(json)))

	diags := resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}
