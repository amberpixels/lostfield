package blankIdent

// ConvertPerformanceMapSchemeModelToProto converts PerformanceMapScheme to PerformanceMapSchemeReply
// Uses blank identifier (_ = ) to acknowledge VenueConfiguration field that is intentionally not converted
func ConvertPerformanceMapSchemeModelToProto(model *PerformanceMapScheme) *PerformanceMapSchemeReply {
	if model == nil {
		return &PerformanceMapSchemeReply{}
	}
	_ = model.VenueConfiguration

	return &PerformanceMapSchemeReply{
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
