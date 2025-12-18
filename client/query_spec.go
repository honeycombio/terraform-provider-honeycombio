package client

import (
	"encoding/json"
	"reflect"
	"time"
)

const (
	DefaultQueryTimeRange = 2 * 60 * 60
	DefaultQueryLimit     = 1000
)

// ValidTimeCompareOffsets are the valid time offsets for comparison queries, in seconds.
var ValidTimeCompareOffsets = []int64{
	int64(30 * time.Minute / time.Second),
	int64(1 * time.Hour / time.Second),
	int64(2 * time.Hour / time.Second),
	int64(8 * time.Hour / time.Second),
	int64(24 * time.Hour / time.Second),
	int64(7 * 24 * time.Hour / time.Second),
	int64(28 * 24 * time.Hour / time.Second),
	int64(182 * 24 * time.Hour / time.Second),
}

// QuerySpec represents a Honeycomb query.
//
// API docs: https://docs.honeycomb.io/api/query-specification/
type QuerySpec struct {
	// ID of a query is only set when QuerySpec is returned from the Queries
	// API. This value should not be set when creating or updating queries.
	ID *string `json:"id,omitempty"`

	// The calculations to return as a time series and summary table. If no
	// calculations are provided, COUNT is applied.
	Calculations []CalculationSpec `json:"calculations,omitempty"`
	// CalculatedFields are temporary Calculated Fields that are
	// created for the query.
	// They are not saved in the dataset and are not available in the dataset
	// after the query is run.
	CalculatedFields []CalculatedFieldSpec `json:"calculated_fields,omitempty"`
	// The filters with which to restrict the considered events.
	Filters []FilterSpec `json:"filters,omitempty"`
	// If multiple filters are specified, filter_combination determines how
	// they are applied. Defaults to AND.
	//
	// From experience it seems the API will never answer with AND, instead
	// always omitting the filter combination field entirely.
	FilterCombination FilterCombination `json:"filter_combination,omitempty"`
	// A list of strings describing the columns by which to break events down
	// into groups.
	Breakdowns []string `json:"breakdowns,omitempty"`
	// A list of objects describing the terms on which to order the query
	// results. Each term must appear in either the breakdowns field or the
	// calculations field.
	Orders []OrderSpec `json:"orders,omitempty"`
	// A list of objects describing filters with which to restrict returned
	// groups. Each column/calculate_op pair must appear in the calculations
	// field. There can be multiple havings for the same column/calculate_op
	// pair.
	Havings []HavingSpec `json:"havings,omitempty"`
	// The maximum number of query results, must be between 1 and 1000.
	Limit *int `json:"limit,omitempty"`
	// The time range of query in seconds. Defaults to two hours. If combined
	// with start time or end time, this time range is added after start time
	// or before end time. Cannot be combined with both start time and end time.
	//
	// For more details, check https://docs.honeycomb.io/api/query-specification/#a-caveat-on-time
	TimeRange *int `json:"time_range,omitempty"`
	// The absolute start time of the query, in Unix Time (= seconds since epoch).
	StartTime *int64 `json:"start_time,omitempty"`
	// The absolute end time of the query, in Unix Time (= seconds since epoch).
	EndTime *int64 `json:"end_time,omitempty"`
	// The time resolution of the query’s graph, in seconds. Valid values are
	// the query’s time range /10 at maximum, and /1000 at minimum.
	Granularity *int `json:"granularity,omitempty"`
	// The time offset for comparison queries, in seconds. Used to compare current
	// time range data with data from a previous time period.
	CompareTimeOffsetSeconds *int `json:"compare_time_offset_seconds,omitempty"`
}

// Encode returns the JSON string representation of the QuerySpec.
func (qs *QuerySpec) Encode() (string, error) {
	b, err := json.Marshal(qs)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Determines if two QuerySpecs are equivalent
func (qs *QuerySpec) EquivalentTo(other QuerySpec) bool {
	// The order of Calculations is important for visualization rendering, so we're looking for equality
	calcMatch := true
	if len(qs.Calculations) != len(other.Calculations) {
		calcMatch = false
	} else {
		for i := range qs.Calculations {
			if !reflect.DeepEqual(qs.Calculations[i], other.Calculations[i]) {
				calcMatch = false
				break
			}
		}
	}
	if !calcMatch {
		// 'COUNT' is the default Calculation and equivalent to an empty Calculations -- check that before we give up
		defaultCalc := []CalculationSpec{{Op: CalculationOpCount}}
		qsC, oC := defaultCalc, defaultCalc
		if len(qs.Calculations) != 0 {
			qsC = qs.Calculations
		}
		if len(other.Calculations) != 0 {
			oC = other.Calculations
		}
		if !reflect.DeepEqual(qsC, oC) {
			return false
		}
	}

	// Orders have a default ascending order, so we need to check equivalence
	if len(qs.Orders) != len(other.Orders) {
		return false
	}
	for i := range qs.Orders {
		if PtrValueOrDefault(qs.Orders[i].Order, SortOrderAsc) != PtrValueOrDefault(other.Orders[i].Order, SortOrderAsc) {
			return false
		}
		if !reflect.DeepEqual(qs.Orders[i].Column, other.Orders[i].Column) {
			return false
		}
		if !reflect.DeepEqual(qs.Orders[i].Op, other.Orders[i].Op) {
			return false
		}
	}

	// the exact order of filters does not matter, but their equvalence does
	if !Equivalent(qs.Filters, other.Filters) {
		return false
	}
	if ValueOrDefault(qs.FilterCombination, DefaultFilterCombination) != ValueOrDefault(other.FilterCombination, DefaultFilterCombination) {
		return false
	}

	if !reflect.DeepEqual(qs.Breakdowns, other.Breakdowns) &&
		// an empty Breakdowns is equivalent to a nil Breakdowns, so we need to check that
		// as DeepEqual will not consider them equal
		((qs.Breakdowns != nil || len(other.Breakdowns) != 0) &&
			(len(qs.Breakdowns) != 0 || other.Breakdowns != nil)) {
		return false
	}

	// the exact order of havings does not matter, but their equvalence does
	if !Equivalent(qs.Havings, other.Havings) {
		return false
	}
	if PtrValueOrDefault(qs.Limit, DefaultQueryLimit) != PtrValueOrDefault(other.Limit, DefaultQueryLimit) {
		return false
	}
	if PtrValueOrDefault(qs.TimeRange, DefaultQueryTimeRange) != PtrValueOrDefault(other.TimeRange, DefaultQueryTimeRange) {
		return false
	}
	if !reflect.DeepEqual(qs.StartTime, other.StartTime) || !reflect.DeepEqual(qs.EndTime, other.EndTime) {
		return false
	}
	// Granularity may be exported out of the Query Builder as '0' when not provided
	if PtrValueOrDefault(qs.Granularity, 0) != PtrValueOrDefault(other.Granularity, 0) {
		return false
	}
	if !reflect.DeepEqual(qs.CompareTimeOffsetSeconds, other.CompareTimeOffsetSeconds) {
		return false
	}

	return true
}

// CalculationSpec represents a calculation within a query.
type CalculationSpec struct {
	Op CalculationOp `json:"op"`
	// Column to perform the operation on. Not needed with COUNT or CONCURRENCY
	Column *string `json:"column,omitempty"`
}

// CalculatedFieldSpec represents a Temporary Calculated Field within a query.
type CalculatedFieldSpec struct {
	Name       string `json:"name"`
	Expression string `json:"expression"`
}

// CalculationOp represents the operator of a calculation.
type CalculationOp string

const (
	CalculationOpCount         CalculationOp = "COUNT"
	CalculationOpConcurrency   CalculationOp = "CONCURRENCY"
	CalculationOpSum           CalculationOp = "SUM"
	CalculationOpAvg           CalculationOp = "AVG"
	CalculationOpCountDistinct CalculationOp = "COUNT_DISTINCT"
	CalculationOpMax           CalculationOp = "MAX"
	CalculationOpMin           CalculationOp = "MIN"
	CalculationOpP001          CalculationOp = "P001"
	CalculationOpP01           CalculationOp = "P01"
	CalculationOpP05           CalculationOp = "P05"
	CalculationOpP10           CalculationOp = "P10"
	CalculationOpP20           CalculationOp = "P20"
	CalculationOpP25           CalculationOp = "P25"
	CalculationOpP50           CalculationOp = "P50"
	CalculationOpP75           CalculationOp = "P75"
	CalculationOpP80           CalculationOp = "P80"
	CalculationOpP90           CalculationOp = "P90"
	CalculationOpP95           CalculationOp = "P95"
	CalculationOpP99           CalculationOp = "P99"
	CalculationOpP999          CalculationOp = "P999"
	CalculationOpHeatmap       CalculationOp = "HEATMAP"
	CalculationOpRateAvg       CalculationOp = "RATE_AVG"
	CalculationOpRateSum       CalculationOp = "RATE_SUM"
	CalculationOpRateMax       CalculationOp = "RATE_MAX"
)

func (c CalculationOp) IsUnaryOp() bool {
	return c == CalculationOpCount || c == CalculationOpConcurrency
}

// CalculationOps returns an exhaustive list of Calculation Operators.
func CalculationOps() []CalculationOp {
	return append(HavingCalculationOps(), CalculationOpHeatmap)
}

// HavingCalculationOps returns an exhaustive list of calculation operators
// supported by Havings. Havings does not support Heatmap.
func HavingCalculationOps() []CalculationOp {
	return []CalculationOp{
		CalculationOpCount,
		CalculationOpConcurrency,
		CalculationOpSum,
		CalculationOpAvg,
		CalculationOpCountDistinct,
		CalculationOpMax,
		CalculationOpMin,
		CalculationOpP001,
		CalculationOpP01,
		CalculationOpP05,
		CalculationOpP10,
		CalculationOpP20,
		CalculationOpP25,
		CalculationOpP50,
		CalculationOpP75,
		CalculationOpP80,
		CalculationOpP90,
		CalculationOpP95,
		CalculationOpP99,
		CalculationOpP999,
		CalculationOpRateAvg,
		CalculationOpRateSum,
		CalculationOpRateMax,
	}
}

// FilterSpec represents a filter within a query.
type FilterSpec struct {
	Column string   `json:"column"`
	Op     FilterOp `json:"op"`
	// Value to use with the filter operation. The type of the filter value
	// depends on the operator:
	//  - 'exists' and 'does-not-exist': value should be nil
	//  - 'in' and 'not-in': value should be a []string
	//  - all other ops: value could be a string, int, bool or float
	Value any `json:"value,omitempty"`
}

// FilterOp represents the operator of a filter.
type FilterOp string

// Declaration of filter operators.
const (
	FilterOpEquals             FilterOp = "="
	FilterOpNotEquals          FilterOp = "!="
	FilterOpGreaterThan        FilterOp = ">"
	FilterOpGreaterThanOrEqual FilterOp = ">="
	FilterOpSmallerThan        FilterOp = "<"
	FilterOpSmallerThanOrEqual FilterOp = "<="
	FilterOpStartsWith         FilterOp = "starts-with"
	FilterOpDoesNotStartWith   FilterOp = "does-not-start-with"
	FilterOpEndsWith           FilterOp = "ends-with"
	FilterOpDoesNotEndWith     FilterOp = "does-not-end-with"
	FilterOpExists             FilterOp = "exists"
	FilterOpDoesNotExist       FilterOp = "does-not-exist"
	FilterOpContains           FilterOp = "contains"
	FilterOpDoesNotContain     FilterOp = "does-not-contain"
	FilterOpIn                 FilterOp = "in"
	FilterOpNotIn              FilterOp = "not-in"
)

// FilterOps returns an exhaustive list of available filter operators.
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
		FilterOpEndsWith,
		FilterOpDoesNotEndWith,
		FilterOpExists,
		FilterOpDoesNotExist,
		FilterOpContains,
		FilterOpDoesNotContain,
		FilterOpIn,
		FilterOpNotIn,
	}
}

// FilterOpFromString converts a string to a FilterOp. Returns an empty FilterOp if the string
// does not match any valid filter operation.
func FilterOpFromString(s string) FilterOp {
	for _, op := range FilterOps() {
		if string(op) == s {
			return op
		}
	}
	return FilterOp("")
}

// IsUnary returns true if the filter operation is unary (does not require a value).
// Unary operations are "exists" and "does-not-exist".
func (f FilterOp) IsUnary() bool {
	return f == FilterOpExists || f == FilterOpDoesNotExist
}

// IsArray returns true if the filter operation requires an array value.
// Array operations are "in" and "not-in".
func (f FilterOp) IsArray() bool {
	return f == FilterOpIn || f == FilterOpNotIn
}

// FilterCombination describes how the filters of a query should be combined.
type FilterCombination string

// Declaration of filter combinations.
const (
	FilterCombinationOr      FilterCombination = "OR"
	FilterCombinationAnd     FilterCombination = "AND"
	DefaultFilterCombination                   = FilterCombinationAnd
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

// HavingSpec describes filters in which to restrict returned groups.
type HavingSpec struct {
	CalculateOp *CalculationOp `json:"calculate_op,omitempty"`
	Column      *string        `json:"column,omitempty"`
	Op          *HavingOp      `json:"op,omitempty"`
	Value       any            `json:"value,omitempty"`
}

// HavingOp represents the operator of a having clause
type HavingOp string

// Declaration of having operations
const (
	HavingOpEquals             HavingOp = "="
	HavingOpNotEquals          HavingOp = "!="
	HavingOpGreaterThan        HavingOp = ">"
	HavingOpGreaterThanOrEqual HavingOp = ">="
	HavingOpLessThan           HavingOp = "<"
	HavingOpLessThanOrEqual    HavingOp = "<="
)

// HavingOps returns an exhaustive list of all having operations.
func HavingOps() []HavingOp {
	return []HavingOp{
		HavingOpEquals,
		HavingOpNotEquals,
		HavingOpGreaterThan,
		HavingOpGreaterThanOrEqual,
		HavingOpLessThan,
		HavingOpLessThanOrEqual,
	}
}
