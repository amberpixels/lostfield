package differentPackages

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
