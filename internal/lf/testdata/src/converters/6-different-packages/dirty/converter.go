package sample_different_packages

import (
	"converters/6-different-packages/models"
)

// MatchedCategoryToProto_Incomplete converts MatchedCategory to MatchedCategoryProto but is incomplete
func MatchedCategoryToProto_Incomplete(category *models.MatchedCategory) *models.MatchedCategoryProto { // want "incomplete converter"
	if category == nil {
		return nil
	}
	return &models.MatchedCategoryProto{
		// Missing: Name field
	}
}

// MatchingDetailsToProto_Incomplete converts MatchingDetails to MatchingDetailsProto but is incomplete
func MatchingDetailsToProto_Incomplete(details models.MatchingDetails) models.MatchingDetailsProto { // want "incomplete converter"
	return models.MatchingDetailsProto{
		// Missing: Info field
	}
}

// MatchedMapDataToProto_Incomplete converts MatchedMapData to MatchedMapDataProto but is incomplete
func MatchedMapDataToProto_Incomplete(data *models.MatchedMapData) *models.MatchedMapDataProto { // want "incomplete converter"
	if data == nil {
		return nil
	}

	categories := make([]models.MatchedCategoryProto, len(data.Categories))
	for i, category := range data.Categories {
		proto := MatchedCategoryToProto_Incomplete(&category)
		if proto != nil {
			categories[i] = *proto
		}
	}

	return &models.MatchedMapDataProto{
		Categories: categories,
		// Missing: Details field
	}
}
