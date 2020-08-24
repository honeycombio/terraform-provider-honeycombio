package honeycombio

// CalculationOpPtr returns a pointer to the given CalculationOp.
func CalculationOpPtr(v CalculationOp) *CalculationOp {
	return &v
}

// IntPtr returns a pointer to the given int.
func IntPtr(v int) *int {
	return &v
}

// Int64Ptr returns a pointer to the given int64.
func Int64Ptr(v int64) *int64 {
	return &v
}

// SortOrderPtr returns a pointer to the given SortOrder.
func SortOrderPtr(v SortOrder) *SortOrder {
	return &v
}

// StringPtr returns a pointer to the given string.
func StringPtr(v string) *string {
	return &v
}
