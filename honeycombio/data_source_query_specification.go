package honeycombio

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/hashcode"
)

func dataSourceHoneycombioQuerySpec() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHoneycombioQuerySpecRead,

		Schema: map[string]*schema.Schema{
			"calculation": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"op": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(helper.AsStringSlice(honeycombio.CalculationOps()), false),
						},
						"column": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"filter": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"column": {
							Type:     schema.TypeString,
							Required: true,
						},
						"op": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(helper.AsStringSlice(honeycombio.FilterOps()), false),
						},
						"value": {
							Type:        schema.TypeString,
							Description: "The value used for the filter. Not needed if op is `exists` or `not-exists`. Mutually exclusive with the other `value_*` options.",
							Optional:    true,
						},
						"value_string": {
							Type:        schema.TypeString,
							Description: "Deprecated: use 'value' instead. The value used for the filter when the column is a string. Mutually exclusive with `value` and the other `value_*` options",
							Optional:    true,
							Deprecated:  "Use of attribute `value_string` is discouraged and will fail to plan if using the empty string. Use of `value` is encouraged.",
						},
						"value_integer": {
							Type:        schema.TypeInt,
							Description: "Deprecated: use 'value' instead. The value used for the filter when the column is an integer. Mutually exclusive with `value` and the other `value_*` options",
							Optional:    true,
							Deprecated:  "Use of attribute `value_integer` is discouraged and will fail to plan if using '0'. Use of `value` is encouraged.",
						},
						"value_float": {
							Type:        schema.TypeFloat,
							Description: "Deprecated: use 'value' instead. The value used for the filter when the column is a float. Mutually exclusive with `value` and the other `value_*` options",
							Optional:    true,
							Deprecated:  "Use of attribute `value_float` is discouraged and will fail to plan if using '0'. Use of `value` is encouraged.",
						},
						"value_boolean": {
							Type:        schema.TypeBool,
							Description: "Deprecated: use 'value' instead. The value used for the filter when the column is a boolean. Mutually exclusive with `value` and the other `value_*` options",
							Optional:    true,
							Deprecated:  "Use of attribute `value_boolean` is discouraged and will fail to plan if using 'false'. Use of `value` is encouraged.",
						},
					},
				},
			},
			"having": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"calculate_op": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(helper.AsStringSlice(honeycombio.HavingCalculationOps()), false),
						},
						"column": {
							Type: schema.TypeString,
							// not required for COUNT
							Optional: true,
						},
						"op": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(helper.AsStringSlice(honeycombio.HavingOps()), false),
						},
						"value": {
							// API currently assumes this is a number
							Type:     schema.TypeFloat,
							Required: true,
						},
					},
				},
			},
			"filter_combination": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      honeycombio.DefaultFilterCombination,
				ValidateFunc: validation.StringInSlice([]string{string(honeycombio.FilterCombinationAnd), string(honeycombio.FilterCombinationOr)}, false),
			},
			"breakdowns": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"order": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"op": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(helper.AsStringSlice(honeycombio.CalculationOps()), false),
						},
						"column": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"order": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(helper.AsStringSlice(honeycombio.SortOrders()), false),
						},
					},
				},
			},
			"limit": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(1, 1000),
			},
			"time_range": {
				Type:     schema.TypeInt,
				Optional: true,
				// Note: this field could have been set to be computed, but
				// when only the JSON output is used by another resource,
				// Terraform isn't able to set the computed value causing a
				// constant diff. By using a default value instead, we don't
				// need the feedback from the API.
				Default: honeycombio.DefaultQueryTimeRange,
			},
			"start_time": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"end_time": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"granularity": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"json": {
				Type:     schema.TypeString,
				Required: false,
				Optional: false,
				Computed: true,
			},
		},
	}
}

func dataSourceHoneycombioQuerySpecRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	calculations, err := extractCalculations(d)
	if err != nil {
		return diag.FromErr(err)
	}

	filters, err := extractFilters(d)
	if err != nil {
		return diag.FromErr(err)
	}

	havings, err := extractHavings(d)
	if err != nil {
		return diag.FromErr(err)
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
			return diag.Errorf("having %d without matching column in query", i)
		}
	}

	// The API doesn't return filter_combination if it's the default
	var filterCombination honeycombio.FilterCombination
	filterString := d.Get("filter_combination").(string)
	if filterString != "" && filterString != string(honeycombio.DefaultFilterCombination) {
		// doing it this way to support possible different filter types in future
		// and having one less place to update them
		filterCombination = honeycombio.FilterCombination(filterString)
	}

	query := &honeycombio.QuerySpec{
		Calculations:      calculations,
		Filters:           filters,
		Havings:           havings,
		FilterCombination: filterCombination,
		Breakdowns:        extractBreakdowns(d),
		Orders:            extractOrders(d),
		Limit:             extractOptionalInt(d, "limit"),
		TimeRange:         extractOptionalInt(d, "time_range"),
		StartTime:         extractOptionalInt64(d, "start_time"),
		EndTime:           extractOptionalInt64(d, "end_time"),
		Granularity:       extractOptionalInt(d, "granularity"),
	}

	if query.TimeRange != nil && query.StartTime != nil && query.EndTime != nil {
		return diag.Errorf("specify at most two of time_range, start_time and end_time")
	}

	if query.TimeRange != nil && query.Granularity != nil {
		if *query.Granularity > (*query.TimeRange / 10) {
			return diag.Errorf("granularity can not be greater than time_range/10")
		}
		if *query.Granularity < (*query.TimeRange / 1000) {
			return diag.Errorf("granularity can not be less than time_range/1000")
		}
	}

	jsonQuery, err := encodeQuery(query)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("json", jsonQuery)
	d.SetId(strconv.Itoa(hashcode.String(jsonQuery)))

	return nil
}

func extractCalculations(d *schema.ResourceData) ([]honeycombio.CalculationSpec, error) {
	calculationSchemas := d.Get("calculation").([]interface{})
	calculations := make([]honeycombio.CalculationSpec, len(calculationSchemas))

	for i := range calculationSchemas {
		calculation := &calculations[i]

		calculation.Op = honeycombio.CalculationOp(d.Get(fmt.Sprintf("calculation.%d.op", i)).(string))

		c, ok := d.GetOk(fmt.Sprintf("calculation.%d.column", i))
		if ok {
			calculations[i].Column = honeycombio.ToPtr(c.(string))
		}

		if calculation.Op.IsUnaryOp() && calculation.Column != nil {
			return nil, fmt.Errorf("calculation op %s should not have an accompanying column", calculation.Op)
		} else if !calculation.Op.IsUnaryOp() && calculation.Column == nil {
			return nil, fmt.Errorf("calculation op %s is missing an accompanying column", calculation.Op)
		}
	}
	// 'COUNT' is the default calculation and will be returned by the API if
	// none have been provided. As this can potentially cause an infinite diff
	// we'll set the default here if we haven't parsed any
	if len(calculations) == 0 {
		calculations = []honeycombio.CalculationSpec{{Op: honeycombio.CalculationOpCount}}
	}

	return calculations, nil
}

func extractFilters(d *schema.ResourceData) ([]honeycombio.FilterSpec, error) {
	filterSchemas := d.Get("filter").([]interface{})
	filters := make([]honeycombio.FilterSpec, len(filterSchemas))

	for i := range filterSchemas {
		honeyFilter, err := extractFilter(d, i)
		if err != nil {
			return nil, err
		}
		filters[i] = honeyFilter
	}
	return filters, nil
}

func extractHavings(d *schema.ResourceData) ([]honeycombio.HavingSpec, error) {
	havingSchemas := d.Get("having").([]interface{})
	havings := make([]honeycombio.HavingSpec, len(havingSchemas))

	for i := range havingSchemas {
		having := &havings[i]

		co, ok := d.GetOk(fmt.Sprintf("having.%d.calculate_op", i))
		if ok {
			having.CalculateOp = honeycombio.ToPtr(honeycombio.CalculationOp(co.(string)))
		}

		c, ok := d.GetOk(fmt.Sprintf("having.%d.column", i))
		if ok {
			having.Column = honeycombio.ToPtr(c.(string))
		}

		op, ok := d.GetOk(fmt.Sprintf("having.%d.op", i))
		if ok {
			having.Op = honeycombio.ToPtr(honeycombio.HavingOp(op.(string)))
		}

		v, ok := d.GetOk(fmt.Sprintf("having.%d.value", i))
		if ok {
			having.Value = v
		}

		if having.CalculateOp.IsUnaryOp() && having.Column != nil {
			return nil, fmt.Errorf("calculate_op %s should not have an accompanying column", *having.CalculateOp)
		}
		if !having.CalculateOp.IsUnaryOp() && having.Column == nil {
			return nil, fmt.Errorf("calculate_op %s requires a column", *having.CalculateOp)
		}
	}

	return havings, nil
}

const multipleValuesError = "must choose one of 'value', 'value_string', 'value_integer', 'value_float', 'value_boolean'"

func extractFilter(d *schema.ResourceData, index int) (honeycombio.FilterSpec, error) {
	var filter honeycombio.FilterSpec

	filter.Column = d.Get(fmt.Sprintf("filter.%d.column", index)).(string)
	filter.Op = honeycombio.FilterOp(d.Get(fmt.Sprintf("filter.%d.op", index)).(string))

	valueSet := false
	v, vOk := d.GetOk(fmt.Sprintf("filter.%d.value", index))
	if vOk {
		filter.Value = coerceValueToType(v.(string))
		valueSet = true
	}
	vs, vsOk := d.GetOk(fmt.Sprintf("filter.%d.value_string", index))
	if vsOk {
		if valueSet {
			return filter, fmt.Errorf(multipleValuesError)
		}
		filter.Value = vs
		valueSet = true
	}
	vi, viOk := d.GetOk(fmt.Sprintf("filter.%d.value_integer", index))
	if viOk {
		if valueSet {
			return filter, fmt.Errorf(multipleValuesError)
		}
		filter.Value = vi
		valueSet = true
	}
	vf, vfOk := d.GetOk(fmt.Sprintf("filter.%d.value_float", index))
	if vfOk {
		if valueSet {
			return filter, fmt.Errorf(multipleValuesError)
		}
		filter.Value = vf
		valueSet = true
	}
	vb, vbOk := d.GetOk(fmt.Sprintf("filter.%d.value_boolean", index))
	if vbOk {
		if valueSet {
			return filter, fmt.Errorf(multipleValuesError)
		}
		filter.Value = vb
	}

	if filter.Op == honeycombio.FilterOpIn || filter.Op == honeycombio.FilterOpNotIn {
		vs, ok := filter.Value.(string)
		if !ok {
			return filter, fmt.Errorf("value must be a string if filter op is 'in' or 'not-in'")
		}
		filter.Value = strings.Split(vs, ",")
	}

	if filter.Op == honeycombio.FilterOpExists || filter.Op == honeycombio.FilterOpDoesNotExist {
		if filter.Value != nil {
			return filter, fmt.Errorf("filter operation %s must not contain a value", filter.Op)
		}
	} else {
		if filter.Value == nil {
			return filter, fmt.Errorf("filter operation %s requires a value", filter.Op)
		}
	}
	return filter, nil
}

func extractBreakdowns(d *schema.ResourceData) []string {
	breakdownsRaw := d.Get("breakdowns").([]interface{})
	breakdowns := make([]string, len(breakdownsRaw))

	for i, b := range breakdownsRaw {
		breakdowns[i] = b.(string)
	}

	return breakdowns
}

func extractOrders(d *schema.ResourceData) []honeycombio.OrderSpec {
	orderSchemas := d.Get("order").([]interface{})
	orders := make([]honeycombio.OrderSpec, len(orderSchemas))

	for i := range orderSchemas {
		order := &orders[i]

		op, ok := d.GetOk(fmt.Sprintf("order.%d.op", i))
		if ok {
			order.Op = honeycombio.ToPtr(honeycombio.CalculationOp(op.(string)))
		}

		c, ok := d.GetOk(fmt.Sprintf("order.%d.column", i))
		if ok {
			order.Column = honeycombio.ToPtr(c.(string))
		}

		so, ok := d.GetOk(fmt.Sprintf("order.%d.order", i))
		if ok {
			ov := honeycombio.SortOrder(so.(string))
			// ascending is the default, API doesn't return or require
			// the field unless value is descending
			//
			// not sending to avoid constant plan diffs
			if ov != honeycombio.SortOrderAsc {
				order.Order = honeycombio.ToPtr(ov)
			}
		}

		// TODO: validation
	}

	return orders
}

func extractOptionalInt(d *schema.ResourceData, key string) *int {
	value, ok := d.GetOk(key)
	if !ok {
		return nil
	}
	return honeycombio.ToPtr(value.(int))
}

func extractOptionalInt64(d *schema.ResourceData, key string) *int64 {
	value, ok := d.GetOk(key)
	if !ok {
		return nil
	}
	return honeycombio.ToPtr(int64(value.(int)))
}
