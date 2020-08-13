package honeycombio

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	honeycombio "github.com/kvrhdn/go-honeycombio"
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

		var column *string
		c := cMap["column"].(string)
		if c != "" {
			column = &c
		}

		calculations[i] = honeycombio.CalculationSpec{
			Op:     honeycombio.CalculationOp(cMap["op"].(string)),
			Column: column,
		}
	}

	filterSchemas := d.Get("filter").(*schema.Set).List()
	filters := make([]honeycombio.FilterSpec, len(filterSchemas))

	for i, f := range filterSchemas {
		fMap := f.(map[string]interface{})

		filters[i] = honeycombio.FilterSpec{
			Column: fMap["column"].(string),
			Op:     honeycombio.FilterOp(fMap["op"].(string)),
			Value:  fMap["value"].(string),
		}

		// Ensure we don't send filter.Value if op is "exists" or
		// "does-not-exist". The Honeycomb API will refuse this.
		//
		// TODO ideally this check is part of the schema (as a ValidateFunc),
		//      but this is not yet supported by the SDK.
		//      https://github.com/hashicorp/terraform-plugin-sdk/issues/155#issuecomment-489699737
		filter := filters[i]
		if filter.Op == honeycombio.FilterOpExists || filter.Op == honeycombio.FilterOpDoesNotExist {
			if filter.Value != "" {
				return diag.Errorf("Filter operation %s must not contain a value", filter.Op)
			}
			filters[i].Value = nil
		}
	}

	filterCombination := honeycombio.FilterCombination(d.Get("filter_combination").(string))

	breakdownsRaw := d.Get("breakdowns").([]interface{})
	breakdowns := make([]string, len(breakdownsRaw))

	for i, b := range breakdownsRaw {
		breakdowns[i] = b.(string)
	}

	query := &honeycombio.QuerySpec{
		Calculations:      calculations,
		Filters:           filters,
		FilterCombination: &filterCombination,
		Breakdowns:        breakdowns,
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
