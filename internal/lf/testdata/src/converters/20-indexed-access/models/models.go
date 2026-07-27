package modelsIndexedAccess

// Cue represents a transcript cue in the domain layer
type Cue struct {
	Start float64
	Text  string
}

// CueDTO represents a transcript cue in the API layer
type CueDTO struct {
	Start float64
	Text  string
}
