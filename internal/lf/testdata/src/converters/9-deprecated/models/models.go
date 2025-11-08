package models

// EventModel represents the input model
type EventModel struct {
	ID string
	// Deprecated: use Name instead
	OldName  string
	Name     string
	Venue    string
	Category string
}

// EventReply represents the output model
type EventReply struct {
	ID string
	// Deprecated: use Name instead
	OldName  string
	Name     string
	Venue    string
	Category string
}
