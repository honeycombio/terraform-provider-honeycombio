package coerce

import (
	"fmt"

	"strconv"
)

// ValueToType converts a string value to its appropriate type.
func ValueToType(i string) interface{} {
	// HCL really has three base types: bool, string, and number
	// The Plugin SDK allows typing a schema field to Int or Float

	// Plugin SDK assumes 64bit so we'll do the same
	if v, err := strconv.ParseInt(i, 10, 64); err == nil {
		return v
	} else if v, err := strconv.ParseFloat(i, 64); err == nil {
		return v
	} else if v, err := strconv.ParseBool(i); err == nil {
		return v
	}
	// fallthrough to string
	return i
}

// ValueToString converts a value of any type to its string representation.
// It handles basic types like string, int, float, and bool.
// For other types (like maps, slices, etc.), it uses fmt.Sprintf to convert to string.
func ValueToString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case int, int32, int64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%f", v)
	case bool:
		return fmt.Sprintf("%t", v)
	default:
		// For other complex types (maps, slices, arrays, etc.), convert to string
		return fmt.Sprintf("%v", v)
	}
}
