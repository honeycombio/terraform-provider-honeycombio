package honeycombio

// Query represents a Honeycomb API Query specification as defined here: https://docs.honeycomb.io/api/query-specification/#calculation-operators
type Query struct {
	Breakdowns        []string      `json:"breakdowns"`
	Calculations      []Calculation `json:"calculations,omitempty"`
	Filters           []Filter      `json:"filters,omitempty"`
	FilterCombination string        `json:"filter_combination,omitempty"`
	Granularity       int           `json:"granularity,omitempty"`
	Orders            []Order       `json:"orders,omitempty"`
	Limit             int           `json:"limit,omitempty"`
	TimeRange         int           `json:"time_range,omitempty"`
	StartTime         int           `json:"start_time,omitempty"`
	EndTime           int           `json:"end_time,omitempty"`
}

type Calculation struct {
	Column string `json:"column,omitempty"`
	Op     string `json:"op"`
}

type Filter struct {
	Column string `json:"column"`
	Op     string `json:"op"`
	Value  string `json:"value,omitempty"`
}

type Order struct {
	Column string `json:"column"`
	Op     string `json:"op"`
	Order  string `json:"orders"`
}
