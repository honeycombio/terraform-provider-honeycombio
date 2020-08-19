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
	//
	// From experience it seems the API will never answer with AND, instead
	// always omitting the filter combination field entirely.
	FilterCombination *FilterCombination `json:"filter_combination,omitempty"`
	// A list of strings describing the columns by which to break events down
	// into groups.
	Breakdowns []string `json:"breakdowns,omitempty"`
	// A list of objects describing the terms on which to order the query
	// results. Each term must appear in either the breakdowns field or the
	// calculations field.
	Orders []OrderSpec `json:"orders,omitempty"`
	// The maximum number of query results, must be between 1 and 1000.
	Limit *int `json:"limit,omitempty"`

	// not all available fields are currently implemented by QuerySpec, see https://docs.honeycomb.io/api/query-specification/#fields-on-a-query-specification
	// fields not implemented yet: granularity, time_range, start_time, end_time
}

// CalculationSpec represents a calculation within a query.
type CalculationSpec struct {
	Op CalculationOp `json:"op"`
	// Column to perform the operation on. Not needed with COUNT.
	Column *string `json:"column,omitempty"`
}

// CalculationOp represents the operation of a calculation.
type CalculationOp string

// Declaration of calculation ops.
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

// CalculationOps returns an exhaustive list of calculation ops.
func CalculationOps() []CalculationOp {
	return []CalculationOp{
		CalculateOpCount,
		CalculateOpSum,
		CalculateOpAvg,
		CalculateOpCountDistinct,
		CalculateOpMax,
		CalculateOpMin,
		CalculateOpP001,
		CalculateOpP01,
		CalculateOpP05,
		CalculateOpP10,
		CalculateOpP25,
		CalculateOpP50,
		CalculateOpP75,
		CalculateOpP90,
		CalculateOpP95,
		CalculateOpP99,
		CalculateOpP999,
		CalculateOpHeatmap,
	}
}

// FilterSpec represents a filter within a query.
type FilterSpec struct {
	Column string   `json:"column"`
	Op     FilterOp `json:"op"`
	// Value to use with the filter operation, not all operations need a value.
	Value interface{} `json:"value,omitempty"`
}

// FilterOp represents the operation of a filter.
type FilterOp string

// Declaration of filter ops.
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

// FilterOps returns an exhaustive list of available filter ops.
func FilterOps() []FilterOp {
	return []FilterOp{
		FilterOpEquals,
		FilterOpNotEquals,
		FilterOpGreaterThan,
		FilterOpGreaterThanOrEqual,
		FilterOpSmallerThan,
		FilterOpSmallerThanOrEqual,
		FilterOpStartsWith,
		FilterOpDoesNotStartWith,
		FilterOpExists,
		FilterOpDoesNotExist,
		FilterOpContains,
		FilterOpDoesNotContain,
	}
}

// FilterCombination describes how the filters of a query should be combined.
type FilterCombination string

// Declaration of filter combinations.
const (
	FilterCombinationOr  FilterCombination = "OR"
	FilterCombinationAnd FilterCombination = "AND"
)

// FilterCombinations returns an exhaustive list of filter combinations.
func FilterCombinations() []FilterCombination {
	return []FilterCombination{FilterCombinationOr, FilterCombinationAnd}
}

// OrderSpec describes how to order the results of a query.
type OrderSpec struct {
	Op     *CalculationOp `json:"op,omitempty"`
	Column *string        `json:"column,omitempty"`
	Order  *SortOrder     `json:"order,omitempty"`
}

// SortOrder describes in which order the results should be sorted.
type SortOrder string

// Declaration of sort orders.
const (
	SortOrderAsc  SortOrder = "ascending"
	SortOrderDesc SortOrder = "descending"
)

// SortOrders returns an exhaustive list of all sort orders.
func SortOrders() []SortOrder {
	return []SortOrder{SortOrderAsc, SortOrderDesc}
}
