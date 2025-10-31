package sliceToNonSlice

// VenueDetail represents input slice element
type VenueDetail struct {
	Name     string
	Sections []string
}

// Metadata represents output non-slice
type Metadata struct {
	Categories []Category
}

// Category represents a category
type Category struct {
	Name     string
	Sections []string
}
