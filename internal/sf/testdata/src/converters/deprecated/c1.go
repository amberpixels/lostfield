package deprecated

// ConvertEventModelToReply converts EventModel to EventReply
// This converter handles all non-deprecated fields.
// OldName is deprecated and is intentionally skipped.
// When ExcludeDeprecated is enabled, no diagnostics are reported.
func ConvertEventModelToReply(model *EventModel) *EventReply {
	if model == nil {
		return &EventReply{}
	}

	return &EventReply{
		OldName:  model.OldName, // Deprecated field - intentionally copied to show we can still use it
		ID:       model.ID,
		Name:     model.Name,
		Venue:    model.Venue,
		Category: model.Category,
	}
}
