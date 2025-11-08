package sample_slice_to_non_slice

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
