package honeycombio

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/kvrhdn/go-honeycombio"
	"github.com/kvrhdn/terraform-provider-honeycombio/util"
)

var validQueryCalculationOps = []string{
	"COUNT",
	"SUM",
	"AVG",
	"COUNT_DISTINCT",
	"MAX",
	"MIN",
	"P001",
	"P01",
	"P05",
	"P10",
	"P25",
	"P50",
	"P75",
	"P90",
	"P95",
	"P99",
	"P999",
	"HEATMAP",
}
var validQueryFilterOps = []string{
	"=",
	"!=",
	">",
	">=",
	"<",
	"<=",
	"starts-with",
	"does-not-start-with",
	"exists",
	"does-not-exist",
	"contains",
	"does-not-contain",
}

func dataSourceHoneycombioQuery() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHoneycombioQueryRead,

		Schema: map[string]*schema.Schema{
			"calculation": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"op": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(validQueryCalculationOps, false),
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
							ValidateFunc: validation.StringInSlice(validQueryFilterOps, false),
						},
						"value": {
							Type:        schema.TypeString,
							Description: "Deprecated: use the explicitly typed `value_string` instead. This variant will potentially break dashboards if used with non-string columns. Mutually exclusive with the other `value_*` options.",
							Optional:    true,
							Deprecated:  "Use of attribute `value` is discouraged and will potentially break dashboards if used with non-string columns. The explicitly typed `value_*` variants are preferred instead.",
						},
						"value_string": {
							Type:        schema.TypeString,
							Description: "The value used for the filter when the column is a string. Mutually exclusive with `value` and the other `value_*` options",
							Optional:    true,
						},
						"value_integer": {
							Type:        schema.TypeInt,
							Description: "The value used for the filter when the column is an integer. Mutually exclusive with `value` and the other `value_*` options",
							Optional:    true,
						},
						"value_float": {
							Type:        schema.TypeFloat,
							Description: "The value used for the filter when the column is a float. Mutually exclusive with `value` and the other `value_*` options",
							Optional:    true,
						},
						"value_boolean": {
							Type:        schema.TypeBool,
							Description: "The value used for the filter when the column is a boolean. Mutually exclusive with `value` and the other `value_*` options",
							Optional:    true,
						},
					},
				},
			},
			"filter_combination": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "AND",
				ValidateFunc: validation.StringInSlice([]string{"AND", "OR"}, false),
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
							ValidateFunc: validation.StringInSlice(validQueryCalculationOps, false),
						},
						"column": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"order": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"ascending", "descending"}, false),
						},
					},
				},
			},
			"limit": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(1, 1000),
			},
			"json": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceHoneycombioQueryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	calculationSchemas := d.Get("calculation").(*schema.Set).List()
	calculations := make([]honeycombio.CalculationSpec, len(calculationSchemas))

	for i, c := range calculationSchemas {
		cMap := c.(map[string]interface{})

		op := honeycombio.CalculationOp(cMap["op"].(string))

		var column *string
		c := cMap["column"].(string)
		if c != "" {
			column = &c
		}

		if op == honeycombio.CalculateOpCount && column != nil {
			return diag.Errorf("calculation op COUNT should not have an accompanying column")
		}

		calculations[i] = honeycombio.CalculationSpec{
			Op:     op,
			Column: column,
		}
	}

	filters, err := extractFilters(d)
	if err != nil {
		return diag.FromErr(err)
	}

	filterCombination := honeycombio.FilterCombination(d.Get("filter_combination").(string))

	breakdownsRaw := d.Get("breakdowns").([]interface{})
	breakdowns := make([]string, len(breakdownsRaw))

	for i, b := range breakdownsRaw {
		breakdowns[i] = b.(string)
	}

	var limit *int
	l := d.Get("limit").(int)
	if l != 0 {
		limit = &l
	}

	orderSchemas := d.Get("order").([]interface{})
	orders := make([]honeycombio.OrderSpec, len(orderSchemas))

	for i, o := range orderSchemas {
		oMap := o.(map[string]interface{})

		var op *honeycombio.CalculationOp
		opValue := honeycombio.CalculationOp(oMap["op"].(string))
		if opValue != "" {
			op = &opValue
		}

		var column *string
		columnValue := oMap["column"].(string)
		if columnValue != "" {
			column = &columnValue
		}

		var sortOrder *honeycombio.SortOrder
		sortOrderValue := honeycombio.SortOrder(oMap["order"].(string))
		if sortOrderValue != "" {
			sortOrder = &sortOrderValue
		}

		orders[i] = honeycombio.OrderSpec{
			Op:     op,
			Column: column,
			Order:  sortOrder,
		}
	}

	query := &honeycombio.QuerySpec{
		Calculations:      calculations,
		Filters:           filters,
		FilterCombination: &filterCombination,
		Breakdowns:        breakdowns,
		Orders:            orders,
		Limit:             limit,
	}

	jsonQuery, err := encodeQuery(query)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("json", jsonQuery)
	d.SetId(strconv.Itoa(util.HashString(jsonQuery)))

	return nil
}

// encodeQuery in a JSON string.
func encodeQuery(q *honeycombio.QuerySpec) (string, error) {
	jsonQueryBytes, err := json.MarshalIndent(q, "", "  ")
	return string(jsonQueryBytes), err
}

type querySpecValidateDiagFunc func(q *honeycombio.QuerySpec) diag.Diagnostics

// validateQueryJSON checks that the input can be deserialized as a QuerySpec
// and optionally runs a list of custom validation functions.
func validateQueryJSON(validators ...querySpecValidateDiagFunc) schema.SchemaValidateDiagFunc {
	return func(i interface{}, path cty.Path) diag.Diagnostics {
		var q honeycombio.QuerySpec

		err := json.Unmarshal([]byte(i.(string)), &q)
		if err != nil {
			return diag.Errorf("Value of query_json is not a valid query specification")
		}

		var diagnostics diag.Diagnostics

		for _, validator := range validators {
			diagnostics = append(diagnostics, validator(&q)...)
		}
		return diagnostics
	}
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

const multipleValuesError = "must choose one of 'value', 'value_string', 'value_integer', 'value_float', 'value_boolean'"

func extractFilter(d *schema.ResourceData, index int) (honeycombio.FilterSpec, error) {
	var filter honeycombio.FilterSpec

	filter.Column = d.Get(fmt.Sprintf("filter.%d.column", index)).(string)
	filter.Op = honeycombio.FilterOp(d.Get(fmt.Sprintf("filter.%d.op", index)).(string))

	valueSet := false
	v, vOk := d.GetOk(fmt.Sprintf("filter.%d.value", index))
	if vOk {
		filter.Value = v
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
		valueSet = true
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
