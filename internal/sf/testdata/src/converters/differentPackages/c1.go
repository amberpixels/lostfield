package differentPackages

// MatchedCategoryToProto converts MatchedCategory to MatchedCategoryProto
func MatchedCategoryToProto(category *MatchedCategory) *MatchedCategoryProto {
	if category == nil {
		return nil
	}
	return &MatchedCategoryProto{
		Name: category.Name,
	}
}

// MatchingDetailsToProto converts MatchingDetails to MatchingDetailsProto
func MatchingDetailsToProto(details MatchingDetails) MatchingDetailsProto {
	return MatchingDetailsProto{
		Info: details.Info,
	}
}

// MatchedMapDataToProto converts MatchedMapData to MatchedMapDataProto
// This has similar type names (MatchedMapData -> MatchedMapDataProto) but should be caught as a converter
// The old logic would incorrectly skip this because of matching name "MatchedMapData"
// The new logic should correctly identify this as a converter because the types are from different packages
func MatchedMapDataToProto(data *MatchedMapData) *MatchedMapDataProto {
	if data == nil {
		return nil
	}

	categories := make([]MatchedCategoryProto, len(data.Categories))
	for i, category := range data.Categories {
		proto := MatchedCategoryToProto(&category)
		if proto != nil {
			categories[i] = *proto
		}
	}

	return &MatchedMapDataProto{
		Categories: categories,
		Details:    MatchingDetailsToProto(data.Details),
	}
}
