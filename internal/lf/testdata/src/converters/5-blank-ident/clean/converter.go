package sample_blank_indent

import (
	"converters/5-blank-ident/models"
)

// ConvertPerformanceMapSchemeModelToProto converts PerformanceMapScheme to PerformanceMapSchemeReply
// Uses blank identifier (_ = ) to acknowledge VenueConfiguration field that is intentionally not converted
func ConvertPerformanceMapSchemeModelToProto(model *models.PerformanceMapScheme) *models.PerformanceMapSchemeReply {
	if model == nil {
		return &models.PerformanceMapSchemeReply{}
	}
	_ = model.VenueConfiguration

	return &models.PerformanceMapSchemeReply{
		MapSlug:            model.MapSlug,
		MapUrl:             model.MapUrl,
		MapUrlFixed:        model.MapUrlFixed,
		MapDecoratedSvgUrl: model.MapDecoratedSvgUrl,
		ThumbPattern:       model.ThumbPattern,
		ThumbFallbackUrl:   model.ThumbFallbackUrl,
		ProviderName:       model.ProviderName,
		ProviderMapRef:     model.ProviderMapRef,
		Categories:         model.Categories,
		Sections:           model.Sections,
	}
}
