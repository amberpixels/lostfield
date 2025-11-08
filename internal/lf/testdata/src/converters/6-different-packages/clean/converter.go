package sample_different_packages

import (
	"converters/6-different-packages/models"
)

// MatchedCategoryToProto converts MatchedCategory to MatchedCategoryProto
func MatchedCategoryToProto(category *models.MatchedCategory) *models.MatchedCategoryProto {
	if category == nil {
		return nil
	}
	return &models.MatchedCategoryProto{
		Name: category.Name,
	}
}

// MatchingDetailsToProto converts MatchingDetails to MatchingDetailsProto
func MatchingDetailsToProto(details models.MatchingDetails) models.MatchingDetailsProto {
	return models.MatchingDetailsProto{
		Info: details.Info,
	}
}

// MatchedMapDataToProto converts MatchedMapData to MatchedMapDataProto
// This has similar type names (MatchedMapData -> MatchedMapDataProto) but should be caught as a converter
// The old logic would incorrectly skip this because of matching name "MatchedMapData"
// The new logic should correctly identify this as a converter because the types are from different packages
func MatchedMapDataToProto(data *models.MatchedMapData) *models.MatchedMapDataProto {
	if data == nil {
		return nil
	}

	categories := make([]models.MatchedCategoryProto, len(data.Categories))
	for i, category := range data.Categories {
		proto := MatchedCategoryToProto(&category)
		if proto != nil {
			categories[i] = *proto
		}
	}

	return &models.MatchedMapDataProto{
		Categories: categories,
		Details:    MatchingDetailsToProto(data.Details),
	}
}
