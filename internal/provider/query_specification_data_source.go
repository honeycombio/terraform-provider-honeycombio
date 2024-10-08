package provider

import (
	"context"
	"reflect"
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
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/hashcode"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/models"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &querySpecDataSource{}
)

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
					"Valid values must be in between the query's time range /10 at maximum, and /1000 at minimum.",
				Optional:   true,
				Validators: []validator.Int64{int64validator.AtLeast(0)},
			},
			"json": schema.StringAttribute{
				Description: "The generated query specification in JSON format.",
				Computed:    true,
				Required:    false,
				Optional:    false,
			},
		},
		Blocks: map[string]schema.Block{
			"calculation": schema.ListNestedBlock{
				Description: "Zero or more configuration blocks describing the calculations to return as a time series and summary table. " +
					"If no calculations are provided, \"COUNT\" is assumed.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"op": schema.StringAttribute{
							Description: "The operatior to apply.",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOf(helper.AsStringSlice(client.CalculationOps())...),
							},
						},
						"column": schema.StringAttribute{
							Description: "The column to apply the operator on. " +
								"Not allowed with \"COUNT\" or \"CONCURRENCY\", required for all other operators.",
							Optional: true,
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
							Description: "The operator to apply.",
							Required:    true,
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
							Description: "The operator to apply.",
							Required:    true,
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
							Description: "The operator to apply.",
							Optional:    true,
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
		calculations = append(calculations, calculation)
	}
	// 'COUNT' is the default calculation and will be returned by the API if
	// none have been provided. As this can potentially cause an infinite diff
	// we'll set the default here if we haven't parsed any
	if len(calculations) == 0 {
		calculations = []client.CalculationSpec{{Op: client.CalculationOpCount}}
	}

	filters := make([]client.FilterSpec, 0, len(data.Filters))
	for i, f := range data.Filters {
		filter := client.FilterSpec{
			Column: f.Column.ValueString(),
			Op:     client.FilterOp(f.Op.ValueString()),
		}

		// TODO: replace with DynamicAttribute
		if !f.Value.IsNull() {
			if filter.Op == client.FilterOpIn || filter.Op == client.FilterOpNotIn {
				// if the filter operation is 'in' or 'not-in' we expect the value
				// to be a CSV string so we build it into a slice
				values := strings.Split(f.Value.ValueString(), ",")
				result := make([]interface{}, len(values))
				for i, value := range values {
					result[i] = coerceValueToType(value)
				}
				filter.Value = result
			} else {
				filter.Value = coerceValueToType(f.Value.ValueString())
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

	querySpec := &client.QuerySpec{
		Calculations:      calculations,
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

	if querySpec.TimeRange != nil && querySpec.StartTime != nil && querySpec.EndTime != nil {
		resp.Diagnostics.AddError(
			"invalid time configuration",
			"specify at most two of time_range, start_time and end_time",
		)
	}
	if querySpec.TimeRange != nil && querySpec.Granularity != nil {
		if *querySpec.Granularity > (*querySpec.TimeRange / 10) {
			resp.Diagnostics.AddAttributeError(
				path.Root("granularity"),
				"invalid granularity",
				"granularity can not be greater than time_range/10",
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

func coerceValueToType(i string) any {
	if v, err := strconv.ParseInt(i, 10, 64); err == nil {
		return v
	} else if v, err := strconv.ParseFloat(i, 64); err == nil {
		return v
	} else if v, err := strconv.ParseBool(i); err == nil {
		return v
	}
	// fallthrough to string
	return i
}
