package models

// MatchedMapData represents input model
type MatchedMapData struct {
	Categories []MatchedCategory
	Details    MatchingDetails
}

// MatchedCategory represents a category in input
type MatchedCategory struct {
	Name string
}

// MatchingDetails represents matching details in input
type MatchingDetails struct {
	Info string
}

// MatchedMapDataProto represents output model
type MatchedMapDataProto struct {
	Categories []MatchedCategoryProto
	Details    MatchingDetailsProto
}

// MatchedCategoryProto represents a category in output
type MatchedCategoryProto struct {
	Name string
}

// MatchingDetailsProto represents matching details in output
type MatchingDetailsProto struct {
	Info string
}
