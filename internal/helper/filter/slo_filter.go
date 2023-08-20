package filter

import (
	"fmt"
	"regexp"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper"
)

type SLODetailFilter struct {
	Type       string
	Value      *string
	ValueRegex *regexp.Regexp
}

func NewDetailSLOFilter(filterType, v, r string) (*SLODetailFilter, error) {
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

	return &SLODetailFilter{
		Type:       filterType,
		Value:      value,
		ValueRegex: valRegexp,
	}, nil
}

func (f *SLODetailFilter) Match(s client.SLO) bool {
	// nil filter fails open
	if f == nil {
		return true
	}
	if f.Value != nil {
		return s.Name == *f.Value
	}
	if f.ValueRegex != nil {
		return f.ValueRegex.MatchString(s.Name)
	}
	return true
}
