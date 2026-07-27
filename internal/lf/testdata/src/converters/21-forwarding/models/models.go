package modelsForwarding

// Location is the domain model.
type Location struct {
	ID    int64
	Title string
}

// LocationDTO is the API model.
type LocationDTO struct {
	ID    int64
	Title string
}
