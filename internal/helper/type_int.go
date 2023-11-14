package helper

// The SLO and burn alert APIs use 'X Per Million' to avoid the problems with floats.
// In the name of offering a nicer UX with percentages, we handle the conversion
// back and forth to allow things like 99.98 to be provided in the HCL and
// handle the conversion to and from 999800

// converts a parts per million value to a floating point percentage
func PPMToFloat(t int) float64 {
	return float64(t) / 10000
}
