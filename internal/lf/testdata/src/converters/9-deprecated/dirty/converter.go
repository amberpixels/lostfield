package sample_deprecated_dirty

// Event represents the input model.
type Event struct {
	ID string
	// Deprecated: use Name instead
	OldName  string
	Name     string
	Venue    string
	Category string
}

// EventReply represents the output model.
type EventReply struct {
	ID string
	// Deprecated: use Name instead
	OldName  string
	Name     string
	Venue    string
	Category string
}

// ConvertEventToReplySkippingDeprecated skips the deprecated OldName field.
// This file is exercised with include-deprecated=true, where deprecated fields
// are validated like any other field, so skipping OldName is reported.
func ConvertEventToReplySkippingDeprecated(model *Event) *EventReply { // want "incomplete converter with missing fields: model.OldName, OldName"
	if model == nil {
		return &EventReply{}
	}

	return &EventReply{
		ID:       model.ID,
		Name:     model.Name,
		Venue:    model.Venue,
		Category: model.Category,
	}
}
