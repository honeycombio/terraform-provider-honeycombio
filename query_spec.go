package honeycombio

// QuerySpec represents a Honeycomb query, as described by https://docs.honeycomb.io/api/query-specification/
type QuerySpec struct {
	// The calculations to return as a time series and summary table. If no
	// calculations are provided, COUNT is applied.
	Calculations []CalculationSpec `json:"calculations,omitempty"`
	// The filters with which to restrict the considered events.
	Filters []FilterSpec `json:"filters,omitempty"`
	// If multiple filters are specified, filter_combination determines how
	// they are applied. Set to OR to match any filter in the filter list,
	// defaults to AND.
	FilterCombination *FilterCombination `json:"filter_combination,omitempty"`
	// A list of strings describing the columns by which to break events down
	// into groups.
	Breakdowns []string `json:"breakdowns,omitempty"`

	// not all available fields are currently implemented by QuerySpec, see https://docs.honeycomb.io/api/query-specification/#fields-on-a-query-specification
}

// CalculationSpec represents a calculation within a query.
type CalculationSpec struct {
	Op CalculationOp `json:"op"`
	// Column to perform the operation on. Should not be used with COUNT, which
	// needs no column.
	Column *string `json:"column,omitempty"`
}

// CalculationOp represents the operation of a calculation.
type CalculationOp string

// List of available calculation op types.
const (
	CalculateOpCount         CalculationOp = "COUNT"
	CalculateOpSum           CalculationOp = "SUM"
	CalculateOpAvg           CalculationOp = "AVG"
	CalculateOpCountDistinct CalculationOp = "COUNT_DISTINCT"
	CalculateOpMax           CalculationOp = "MAX"
	CalculateOpMin           CalculationOp = "MIN"
	CalculateOpP001          CalculationOp = "P001"
	CalculateOpP01           CalculationOp = "P01"
	CalculateOpP05           CalculationOp = "P05"
	CalculateOpP10           CalculationOp = "P10"
	CalculateOpP25           CalculationOp = "P25"
	CalculateOpP50           CalculationOp = "P50"
	CalculateOpP75           CalculationOp = "P75"
	CalculateOpP90           CalculationOp = "P90"
	CalculateOpP95           CalculationOp = "P95"
	CalculateOpP99           CalculationOp = "P99"
	CalculateOpP999          CalculationOp = "P999"
	CalculateOpHeatmap       CalculationOp = "HEATMAP"
)

// FilterSpec represents a filter within a query.
type FilterSpec struct {
	Column string   `json:"column"`
	Op     FilterOp `json:"op"`
	// Value to use with the filter operation, not all operations need a value.
	Value interface{} `json:"value,omitempty"`
}

// FilterOp represents the operation of a filter.
type FilterOp string

// List of available filter op types.
const (
	FilterOpEquals             FilterOp = "="
	FilterOpNotEquals          FilterOp = "!="
	FilterOpGreaterThan        FilterOp = ">"
	FilterOpGreaterThanOrEqual FilterOp = ">="
	FilterOpSmallerThan        FilterOp = "<"
	FilterOpSmallerThanOrEqual FilterOp = "<="
	FilterOpStartsWith         FilterOp = "starts-with"
	FilterOpDoesNotStartWith   FilterOp = "does-not-start-with"
	FilterOpExists             FilterOp = "exists"
	FilterOpDoesNotExist       FilterOp = "does-not-exist"
	FilterOpContains           FilterOp = "contains"
	FilterOpDoesNotContain     FilterOp = "does-not-contain"
)

// FilterCombination describes how the filters of a query should be combined.
type FilterCombination string

// List of available filter combination options.
const (
	FilterCombinationOr  FilterCombination = "OR"
	FilterCombinationAnd FilterCombination = "AND"
)
