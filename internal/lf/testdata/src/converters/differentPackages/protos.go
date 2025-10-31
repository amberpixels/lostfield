package differentPackages

// Note: In real scenario, these would be in a different package (pb/proto generated code)
// For testing purposes, we'll simulate them as separate types with similar names

// MatchedMapDataProto represents output model (simulating pbVenueConfig.MatchedMapData)
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
