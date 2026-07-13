package sample_deprecated

// Note: models live in the same package as the converter because deprecated-field
// detection reads doc comments from the ASTs of the current analysis pass only.
// Fields of types imported from other packages cannot be identified as deprecated.

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

// ConvertEventToReply converts Event to EventReply.
// The deprecated OldName field is intentionally skipped: by default
// (include-deprecated=false) deprecated fields are excluded from validation,
// so no diagnostics are reported.
func ConvertEventToReply(model *Event) *EventReply {
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

// ConvertEventToReplyFull copies every field including the deprecated one.
// Using a deprecated field is always allowed - it just isn't required.
func ConvertEventToReplyFull(model *Event) *EventReply {
	if model == nil {
		return &EventReply{}
	}

	return &EventReply{
		OldName:  model.OldName, // Deprecated field - still fine to copy it
		ID:       model.ID,
		Name:     model.Name,
		Venue:    model.Venue,
		Category: model.Category,
	}
}
