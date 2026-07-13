package sample_aggregating

// VenueDetail represents an input slice element.
type VenueDetail struct {
	Name     string
	Sections []string
}

// Metadata is the non-slice output wrapping the aggregated slice.
type Metadata struct {
	Categories []Category
}

// Category is the output slice element type.
type Category struct {
	Name     string
	Sections []string
}

// AggregateDetailsIncomplete is an aggregating converter (slice -> non-slice)
// that drops Sections on both sides: detail.Sections is never read and
// Category.Sections is never set.
func AggregateDetailsIncomplete(details []*VenueDetail) Metadata { // want "incomplete converter with missing fields: detail.Sections, Categories\\[\\].Sections"
	categories := make([]Category, 0, len(details))

	for _, detail := range details {
		if detail == nil {
			continue
		}
		category := Category{
			Name: detail.Name,
		}
		categories = append(categories, category)
	}

	return Metadata{
		Categories: categories,
	}
}
