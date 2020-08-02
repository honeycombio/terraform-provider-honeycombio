package honeycombio

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	honeycombio "github.com/kvrhdn/go-honeycombio"
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

type querySpecConstraints struct {
	calculationCount int
}

func defaultQuerySpecConstraints() *querySpecConstraints {
	return &querySpecConstraints{
		calculationCount: 0,
	}
}

func createQuerySpecSchema(constraints *querySpecConstraints) *schema.Schema {
	if constraints == nil {
		constraints = defaultQuerySpecConstraints()
	}

	return &schema.Schema{
		Type:     schema.TypeSet,
		Required: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"calculation": {
					Type:     schema.TypeSet,
					Required: true,
					MinItems: constraints.calculationCount,
					MaxItems: constraints.calculationCount,
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
					Type:     schema.TypeSet,
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
								Type:     schema.TypeString,
								Optional: true,
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
			},
		},
	}
}

func expandQuerySpec(s *schema.Set) *honeycombio.QuerySpec {
	d := s.List()[0].(map[string]interface{})

	calculationSchema := d["calculation"].(*schema.Set).List()
	calculations := make([]honeycombio.CalculationSpec, len(calculationSchema))

	for i, c := range calculationSchema {
		cMap := c.(map[string]interface{})

		column := cMap["column"].(string)

		calculations[i] = honeycombio.CalculationSpec{
			Op:     honeycombio.CalculationOp(cMap["op"].(string)),
			Column: &column,
		}
	}

	filterSchemas := d["filter"].(*schema.Set).List()
	var filters []honeycombio.FilterSpec

	for _, filterSchema := range filterSchemas {
		filters = append(filters, honeycombio.FilterSpec{
			Column: filterSchema.(map[string]interface{})["column"].(string),
			Op:     honeycombio.FilterOp(filterSchema.(map[string]interface{})["op"].(string)),
			Value:  filterSchema.(map[string]interface{})["value"].(string),
		})
	}

	var filterCombination *honeycombio.FilterCombination

	filterCombinationRaw, ok := d["filter_combination"]
	if ok {
		value := honeycombio.FilterCombination(filterCombinationRaw.(string))
		filterCombination = &value
	} else {
		filterCombination = nil
	}

	breakdownsRaw := d["breakdowns"].([]interface{})
	breakdowns := make([]string, len(breakdownsRaw))

	for i, raw := range breakdownsRaw {
		breakdowns[i] = raw.(string)
	}

	return &honeycombio.QuerySpec{
		Calculations:      calculations,
		Filters:           filters,
		FilterCombination: filterCombination,
		Breakdowns:        breakdowns,
	}
}

func flattenQuerySpec(q *honeycombio.QuerySpec) []map[string]interface{} {
	return []map[string]interface{}{
		{
			"calculation":        flattenCalculations(q.Calculations),
			"filter":             flattenFilters(q.Filters),
			"filter_combination": string(*q.FilterCombination),
			"breakdowns":         q.Breakdowns,
		},
	}
}

func flattenCalculations(cs []honeycombio.CalculationSpec) []map[string]interface{} {
	result := make([]map[string]interface{}, len(cs))

	for i, c := range cs {
		result[i] = map[string]interface{}{
			"op":     c.Op,
			"column": c.Column,
		}
	}

	return result
}

func flattenFilters(fs []honeycombio.FilterSpec) []map[string]interface{} {
	result := make([]map[string]interface{}, len(fs))

	for i, f := range fs {
		result[i] = map[string]interface{}{
			"column": f.Column,
			"op":     f.Op,
			"value":  f.Value,
		}
	}

	return result
}
