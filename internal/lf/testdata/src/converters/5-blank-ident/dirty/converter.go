package sample_blank_indent

import (
	"converters/5-blank-ident/models"
)

// ConvertPerformanceMapSchemeModelToProto_WithoutBlankIdent converts PerformanceMapScheme to PerformanceMapSchemeReply
// but WITHOUT using blank identifier to acknowledge the intentionally skipped field
// This should report VenueConfiguration as missing
func ConvertPerformanceMapSchemeModelToProto_WithoutBlankIdent(model *models.PerformanceMapScheme) *models.PerformanceMapSchemeReply { // want "incomplete converter"
	if model == nil {
		return &models.PerformanceMapSchemeReply{}
	}
	// NOT acknowledging VenueConfiguration with blank ident

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
