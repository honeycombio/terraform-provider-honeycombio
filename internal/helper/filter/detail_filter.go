package filter

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// DetailFilter provides filtering capabilities for resources.
type DetailFilter struct {
	Field      string
	Operator   string
	Value      string
	ValueRegex *regexp.Regexp
}

func NewDetailFilter(field, operator, value, regex string) (*DetailFilter, error) {
	var valRegex *regexp.Regexp
	if regex != "" {
		var err error
		valRegex, err = regexp.Compile(regex)
		if err != nil {
			return nil, fmt.Errorf("invalid regex: %v", err)
		}
	}

	return &DetailFilter{
		Field:      field,
		Operator:   operator,
		Value:      value,
		ValueRegex: valRegex,
	}, nil
}

// Match determines if a resource matches the filter criteria
func (f *DetailFilter) Match(resource interface{}) bool {
	if f == nil {
		return true
	}

	fieldValue, found := getFieldValue(resource, f.Field)
	if !found {
		return false
	}

	strValue := convertToString(fieldValue)

	return compareValues(strValue, f.Operator, f.Value, f.ValueRegex)
}

// getFieldValue extracts a field value from a resource using reflection
// Field name matching is always case-insensitive
func getFieldValue(resource interface{}, fieldName string) (interface{}, bool) {
	v := reflect.ValueOf(resource)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	lowerFieldName := strings.ToLower(fieldName)

	switch v.Kind() {
	case reflect.Struct:
		t := v.Type()
		for i := range t.NumField() {
			if strings.ToLower(t.Field(i).Name) == lowerFieldName {
				return v.Field(i).Interface(), true
			}
		}
		return nil, false
	case reflect.Map:
		for _, key := range v.MapKeys() {
			// Convert key to string for comparison
			keyStr, ok := key.Interface().(string)
			if !ok {
				continue
			}

			if strings.ToLower(keyStr) == lowerFieldName {
				return v.MapIndex(key).Interface(), true
			}
		}
		return nil, false
	default:
		return nil, false
	}
}

// convertToString converts any value to its string representation
func convertToString(value interface{}) string {
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

// compareValues compares two values using the specified operator
func compareValues(strValue, operator, filterValue string, regex *regexp.Regexp) bool {
	switch operator {
	case "equals": // Default to equals if no operator specified
		return strValue == filterValue
	case "not_equals":
		return strValue != filterValue
	case "contains":
		return strings.Contains(strValue, filterValue)
	case "starts_with":
		resp := strings.HasPrefix(strValue, filterValue)
		return resp
	case "ends_with":
		return strings.HasSuffix(strValue, filterValue)
	case "greater_than":
		// Try to convert both to numbers for comparison
		val1, err1 := strconv.ParseFloat(strValue, 64)
		val2, err2 := strconv.ParseFloat(filterValue, 64)
		if err1 == nil && err2 == nil {
			return val1 > val2
		}
		return false
	case "less_than":
		val1, err1 := strconv.ParseFloat(strValue, 64)
		val2, err2 := strconv.ParseFloat(filterValue, 64)
		if err1 == nil && err2 == nil {
			return val1 < val2
		}
		return false
	default:
		if regex != nil {
			return regex.MatchString(strValue)
		}
		return false
	}
}

// MatchName checks if the filter matches a given name.
func (f *DetailFilter) MatchName(name string) bool {
	// nil filter fails open
	if f == nil {
		return true
	}

	if f.Field == "name" {
		if f.ValueRegex != nil {
			return f.ValueRegex.MatchString(name)
		}
		return f.Value == "" || f.Value == name
	}

	// Otherwise, this filter doesn't apply to name matching
	return true
}
