package filter

import (
	"fmt"
	"regexp"

	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper"
)

type DetailFilter struct {
	Type       string
	Value      *string
	ValueRegex *regexp.Regexp
}

func NewDetailFilter(filterType, v, r string) (*DetailFilter, error) {
	if filterType != "name" {
		return nil, fmt.Errorf("only name is supported as a filter type")
	}
	if v != "" && r != "" {
		return nil, fmt.Errorf("only one of value or value_regex may be provided")
	}
	if v == "" && r == "" {
		return nil, fmt.Errorf("one of value or value_regex must be provided")
	}

	var value *string
	var valRegexp *regexp.Regexp
	if v != "" {
		value = helper.ToPtr(v)
	}
	if r != "" {
		valRegexp = regexp.MustCompile(r)
	}

	return &DetailFilter{
		Type:       filterType,
		Value:      value,
		ValueRegex: valRegexp,
	}, nil
}

func (f *DetailFilter) MatchName(name string) bool {
	// nil filter fails open
	if f == nil {
		return true
	}
	if f.Value != nil {
		return name == *f.Value
	}
	if f.ValueRegex != nil {
		return f.ValueRegex.MatchString(name)
	}
	return true
}
