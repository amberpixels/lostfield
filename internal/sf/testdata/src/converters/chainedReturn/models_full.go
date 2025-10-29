package chainedReturn

// VenueConfig represents input proto
type VenueConfig struct {
	ID        int
	Name      string
	Deprecated bool
}

// VenueModel represents output model
type VenueModel struct {
	ID         int
	Name       string
	CreatedAt  string
	UpdatedAt  string
	IsDeprecated bool
	MapSlug    string
	Priority   int
}
