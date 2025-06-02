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

	// Create alternate version with underscores removed for camelCase comparison
	// Eg. "time_period_days" becomes "timeperioddays"
	noUnderscoreFieldName := strings.ReplaceAll(lowerFieldName, "_", "")

	switch v.Kind() {
	case reflect.Struct:
		t := v.Type()
		for i := range t.NumField() {
			fieldType := t.Field(i)
			fieldNameLower := strings.ToLower(fieldType.Name)

			// Direct match
			if fieldNameLower == lowerFieldName {
				return v.Field(i).Interface(), true
			}

			// Match with underscores removed (camelCase to snake_case conversion)
			if strings.ReplaceAll(fieldNameLower, "_", "") == noUnderscoreFieldName {
				return v.Field(i).Interface(), true
			}
		}
		return nil, false
	case reflect.Map:
		for _, key := range v.MapKeys() {
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
	if regex != nil {
		return regex.MatchString(strValue)
	}

	switch operator {
	case "equals", "=", "eq", "":
		return strValue == filterValue
	case "not_equals", "!=", "ne":
		return strValue != filterValue
	case "contains", "in":
		return strings.Contains(strValue, filterValue)
	case "does-not-contain", "not-in":
		return !strings.Contains(strValue, filterValue)
	case "starts_with":
		resp := strings.HasPrefix(strValue, filterValue)
		return resp
	case "does-not-start-with":
		return !strings.HasPrefix(strValue, filterValue)
	case "ends_with":
		return strings.HasSuffix(strValue, filterValue)
	case "does-not-end-with":
		return !strings.HasSuffix(strValue, filterValue)
	case "greater_than", ">", "gt":
		// Try to convert both to numbers for comparison
		val1, err1 := strconv.ParseFloat(strValue, 64)
		val2, err2 := strconv.ParseFloat(filterValue, 64)
		if err1 == nil && err2 == nil {
			return val1 > val2
		}
		return false
	case "greater_than_or_equal", ">=", "ge":
		val1, err1 := strconv.ParseFloat(strValue, 64)
		val2, err2 := strconv.ParseFloat(filterValue, 64)
		if err1 == nil && err2 == nil {
			return val1 >= val2
		}
		return false
	case "less_than", "<", "lt":
		val1, err1 := strconv.ParseFloat(strValue, 64)
		val2, err2 := strconv.ParseFloat(filterValue, 64)
		if err1 == nil && err2 == nil {
			return val1 < val2
		}
		return false
	case "less_than_or_equal", "<=", "le":
		val1, err1 := strconv.ParseFloat(strValue, 64)
		val2, err2 := strconv.ParseFloat(filterValue, 64)
		if err1 == nil && err2 == nil {
			return val1 <= val2
		}
		return false
	case "does-not-exist":
		return strValue == ""
	default:
		return false
	}
}

// MatchName checks if the filter matches a given name.
func (f *DetailFilter) MatchName(name string) bool {
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
