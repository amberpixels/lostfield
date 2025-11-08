package sample_same_type

// applyFilters applies filter conditions to the query
// This is NOT a converter - it takes *MockDB and returns *MockDB (same type)
// The linter should NOT flag this as a converter
func applyFilters(query *MockDB, options QueryOptions) *MockDB {
	// Filter by IDs (OR logic)
	if len(options.IDs) == 1 {
		query.Query += "id = " + options.IDs[0]
	} else if len(options.IDs) > 1 {
		query.Query += "id IN (...)"
	}

	// Filter by name
	if options.Name != "" {
		query.Query += "name = " + options.Name
	}

	return query
}
