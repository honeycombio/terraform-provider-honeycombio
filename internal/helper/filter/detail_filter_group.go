package filter

// FilterGroup represents a group of filters that are combined with AND logic
type FilterGroup struct {
	Filters []*DetailFilter
}

// NewFilterGroup creates a new filter group with the provided filters
func NewFilterGroup(filters []*DetailFilter) *FilterGroup {
	return &FilterGroup{
		Filters: filters,
	}
}

// Match determines if all filters in the group match the resource
// TODO: Implement OR logic if needed in the future
func (g *FilterGroup) Match(resource interface{}) bool {
	if g == nil || len(g.Filters) == 0 {
		return true
	}

	// All filters must match (AND logic)
	for _, filter := range g.Filters {
		if !filter.Match(resource) {
			return false
		}
	}

	return true
}
