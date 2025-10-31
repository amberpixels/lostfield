package sliceToNonSlice

// ProtoDetailsToMetadata converts slice of VenueDetail to single Metadata
// Input: []*VenueDetail (slice)
// Output: Metadata (single struct)
// This is slice->non-slice which should NOT be caught as a converter
func ProtoDetailsToMetadata(details []*VenueDetail) Metadata {
	categories := make([]Category, 0, len(details))

	for _, detail := range details {
		if detail == nil {
			continue
		}

		category := Category{
			Name:     detail.Name,
			Sections: detail.Sections,
		}
		categories = append(categories, category)
	}

	return Metadata{
		Categories: categories,
	}
}
