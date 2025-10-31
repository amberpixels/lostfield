package sameType

// QueryOptions represents query options
type QueryOptions struct {
	IDs []string
	Name string
}

// MockDB represents a mock database connection
type MockDB struct {
	Query string
}
