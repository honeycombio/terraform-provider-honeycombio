package honeycombio

import (
	"strconv"

	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
)

func boardStyleStrings() []string {
	in := honeycombio.BoardStyles()
	out := make([]string, len(in))

	for i := range in {
		out[i] = string(in[i])
	}

	return out
}

func boardQueryStyleStrings() []string {
	in := honeycombio.BoardQueryStyles()
	out := make([]string, len(in))

	for i := range in {
		out[i] = string(in[i])
	}

	return out
}

func calculationOpStrings() []string {
	in := honeycombio.CalculationOps()
	out := make([]string, len(in))

	for i := range in {
		out[i] = string(in[i])
	}

	return out
}

func columnTypeStrings() []string {
	in := honeycombio.ColumnTypes()
	out := make([]string, len(in))

	for i := range in {
		out[i] = string(in[i])
	}

	return out
}

func filterOpStrings() []string {
	in := honeycombio.FilterOps()
	out := make([]string, len(in))

	for i := range in {
		out[i] = string(in[i])
	}

	return out
}

func havingOpStrings() []string {
	in := honeycombio.HavingOps()
	out := make([]string, len(in))

	for i := range in {
		out[i] = string(in[i])
	}

	return out
}

func havingCalculateOpStrings() []string {
	in := honeycombio.CalculationOps()
	out := make([]string, len(in))

	for i := range in {
		// havings cannot use HEATMAP
		if in[i] != honeycombio.CalculationOpHeatmap {
			out[i] = string(in[i])
		}
	}

	return out
}

func sortOrderStrings() []string {
	in := honeycombio.SortOrders()
	out := make([]string, len(in))

	for i := range in {
		out[i] = string(in[i])
	}

	return out
}

func triggerRecipientTypeStrings() []string {
	in := honeycombio.TriggerRecipientTypes()
	out := make([]string, len(in))

	for i := range in {
		out[i] = string(in[i])
	}

	return out
}

func triggerThresholdOpStrings() []string {
	in := honeycombio.TriggerThresholdOps()
	out := make([]string, len(in))

	for i := range in {
		out[i] = string(in[i])
	}

	return out
}

func coerceValueToType(i string) interface{} {
	// HCL really has three base types: bool, string, and number
	// The Plugin SDK allows typing a schema field to Int or Float

	// Plugin SDK assumes 64bit so we'll do the same
	if v, err := strconv.ParseInt(i, 10, 64); err == nil {
		return int(v)
	} else if v, err := strconv.ParseFloat(i, 64); err == nil {
		return v
	} else if v, err := strconv.ParseBool(i); err == nil {
		return v
	}
	// fallthrough to string
	return i
}

// The SLO API uses 'Target Per Million' to avoid the problems with floats.
// In the name of offering a nicer UX with percentages, we handle the conversion
// back and fourth to allow things like 99.98 to be provided in the HCL and
// handle the conversion to and from 999800

// converts a floating point percentage to a 'Target Per Million' SLO value
func floatToTPM(f float64) int {
	return int(f * 10000)
}

// converts a SLO 'Target Per Million' value to a floating point percentage
func tpmToFloat(t int) float64 {
	return float64(t) / 10000
}
