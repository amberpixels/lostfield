package blankIdent

// PerformanceMapScheme represents the input model
type PerformanceMapScheme struct {
	MapSlug            string
	MapUrl             string
	MapUrlFixed        string
	MapDecoratedSvgUrl string
	ThumbPattern       string
	ThumbFallbackUrl   string
	ProviderName       string
	ProviderMapRef     string
	Categories         []string
	Sections           []string
	VenueConfiguration string // This field should be acceptable to mark with _
}

// PerformanceMapSchemeReply represents the output model
type PerformanceMapSchemeReply struct {
	MapSlug            string
	MapUrl             string
	MapUrlFixed        string
	MapDecoratedSvgUrl string
	ThumbPattern       string
	ThumbFallbackUrl   string
	ProviderName       string
	ProviderMapRef     string
	Categories         []string
	Sections           []string
}
