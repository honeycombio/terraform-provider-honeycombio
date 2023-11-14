package helper

import (
	"fmt"
)

// The SLO and burn alert APIs use 'X Per Million' to avoid the problems with floats.
// In the name of offering a nicer UX with percentages, we handle the conversion
// back and forth to allow things like 99.98 to be provided in the HCL and
// handle the conversion to and from 999800

// FloatToPPM converts a floating point percentage to a parts per million value
func FloatToPPM(f float64) int {
	return int(f * 10000)
}

// FloatToPercentString converts a floating point percentage to a string, with a maximum of
// 4 decimal places, no trailing zeros
func FloatToPercentString(f float64) string {
	return fmt.Sprintf("%g", f)
}
