package filter

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/coerce"
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

	strValue := coerce.ValueToString(fieldValue)

	// For tag fields, convert to "key:value" format strings before comparison
	// eg. "tags" field with value {"env": "prod", "team": "ops"} becomes "env:prod,team:ops"
	if strings.ToLower(f.Field) == "tags" {
		strValue = formatTagsAsString(fieldValue)
	}

	return compareValues(strValue, f.Operator, f.Value, f.ValueRegex)
}

// formatTagsAsString converts tag fields to a string representation in "key:value" format
func formatTagsAsString(tagField interface{}) string {
	v := reflect.ValueOf(tagField)
	if v.Kind() == reflect.Ptr && !v.IsNil() {
		v = v.Elem()
	}

	var tagPairs []string

	switch v.Kind() {
	// Tags should typically come back as a slice but adding support for maps
	// just in case.
	case reflect.Map:
		for _, key := range v.MapKeys() {
			keyStr := fmt.Sprintf("%v", key.Interface())
			valueStr := fmt.Sprintf("%v", v.MapIndex(key).Interface())
			tagPairs = append(tagPairs, fmt.Sprintf("%s:%s", keyStr, valueStr))
		}
	case reflect.Slice, reflect.Array:
		for i := range v.Len() {
			item := v.Index(i).Interface()

			// Check if it's a string (simple tag)
			if tagStr, ok := item.(string); ok {
				tagPairs = append(tagPairs, tagStr)
				continue
			}

			itemValue := reflect.ValueOf(item)
			if itemValue.Kind() == reflect.Struct {
				keyField := itemValue.FieldByName("Key")
				valueField := itemValue.FieldByName("Value")

				if keyField.IsValid() && valueField.IsValid() {
					key := fmt.Sprintf("%v", keyField.Interface())
					value := fmt.Sprintf("%v", valueField.Interface())
					tagPairs = append(tagPairs, fmt.Sprintf("%s:%s", key, value))
				}
			}
		}
	}

	// As a fallback, if no tags were found, return the string representation of the field
	if len(tagPairs) == 0 {
		return coerce.ValueToString(tagField)
	}

	return strings.Join(tagPairs, ",")
}

// getFieldValue extracts a field value from a resource using reflection
// Field name matching is always case-insensitive
func getFieldValue(resource interface{}, fieldName string) (interface{}, bool) {
	v := reflect.ValueOf(resource)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	lowerFieldName := strings.ToLower(fieldName)

	// Create alternate version without underscores.
	// This is useful for fields that might be named in camelCase or snake_case.
	// This can occur when resources are read as structs or maps. The struct fields
	// are typically in camelCase, while the filter field names might be in snake_case.
	//
	// For example, "timePeriodDays" becomes "timeperioddays" and the filter might use
	// "time_period_days" or "timeperioddays". Underscores are removed from the field name
	// to allow matching both camelCase and snake_case styles.
	// This allows matching both styles.
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

// compareValues compares two values using the specified operator
func compareValues(strValue, operator, filterValue string, regex *regexp.Regexp) bool {
	if regex != nil {
		return regex.MatchString(strValue)
	}

	switch operator {
	case "equals", "=", "eq", "":
		return strValue == filterValue
	case "not-equals", "!=", "ne":
		return strValue != filterValue
	case "contains", "in":
		return strings.Contains(strValue, filterValue)
	case "does-not-contain", "not-in":
		return !strings.Contains(strValue, filterValue)
	case "starts-with":
		return strings.HasPrefix(strValue, filterValue)
	case "does-not-start-with":
		return !strings.HasPrefix(strValue, filterValue)
	case "ends-with":
		return strings.HasSuffix(strValue, filterValue)
	case "does-not-end-with":
		return !strings.HasSuffix(strValue, filterValue)
	case ">", "gt":
		// Try to convert both to numbers for comparison
		val1, err1 := strconv.ParseFloat(strValue, 64)
		val2, err2 := strconv.ParseFloat(filterValue, 64)
		if err1 == nil && err2 == nil {
			return val1 > val2
		}
		return false
	case ">=", "ge":
		val1, err1 := strconv.ParseFloat(strValue, 64)
		val2, err2 := strconv.ParseFloat(filterValue, 64)
		if err1 == nil && err2 == nil {
			return val1 >= val2
		}
		return false
	case "<", "lt":
		val1, err1 := strconv.ParseFloat(strValue, 64)
		val2, err2 := strconv.ParseFloat(filterValue, 64)
		if err1 == nil && err2 == nil {
			return val1 < val2
		}
		return false
	case "<=", "le":
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
