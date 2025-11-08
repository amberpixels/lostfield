package sample_deprecated

import (
	"converters/9-deprecated/models"
)

// ConvertEventModelToReply converts EventModel to EventReply
// This converter handles all non-deprecated fields.
// OldName is deprecated and is intentionally skipped.
// When ExcludeDeprecated is enabled, no diagnostics are reported.
func ConvertEventModelToReply(model *models.EventModel) *models.EventReply {
	if model == nil {
		return &models.EventReply{}
	}

	return &models.EventReply{
		OldName:  model.OldName, // Deprecated field - intentionally copied to show we can still use it
		ID:       model.ID,
		Name:     model.Name,
		Venue:    model.Venue,
		Category: model.Category,
	}
}
