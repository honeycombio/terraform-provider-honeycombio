package honeycombio

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
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

func dataSourceHoneycombioQuery() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceHoneycombioQueryRead,

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
						//TODO add validation to make sure this doesn't exist when Op in ('exists', 'does-not-exist') if v2 SDK supports that
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

func dataSourceHoneycombioQueryRead(d *schema.ResourceData, meta interface{}) error {
	calculationSchemas := d.Get("calculation").(*schema.Set).List()
	calculations := make([]honeycombio.CalculationSpec, len(calculationSchemas))

	for i, c := range calculationSchemas {
		cMap := c.(map[string]interface{})

		column := cMap["column"].(string)

		calculations[i] = honeycombio.CalculationSpec{
			Op:     honeycombio.CalculationOp(cMap["op"].(string)),
			Column: &column,
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
		hf := filters[i]
		if hf.Op == honeycombio.FilterOpExists || hf.Op == honeycombio.FilterOpDoesNotExist {
			if hf.Value != "" {
				return errors.New(fmt.Sprintf("Filter operation %s must not contain a value", hf.Op))
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
		return err
	}

	d.Set("json", jsonQuery)
	d.SetId(strconv.Itoa(hashcode.String(jsonQuery)))

	return nil
}

// encodeQuery in a JSON string.
func encodeQuery(q *honeycombio.QuerySpec) (string, error) {
	jsonQueryBytes, err := json.MarshalIndent(q, "", "  ")
	return string(jsonQueryBytes), err
}
