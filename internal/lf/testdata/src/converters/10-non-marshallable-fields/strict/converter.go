package strict

import "converters/10-non-marshallable-fields/models"

// In "strict" mode, the Handler field must be handled/read
// This converter is incomplete - missing Handler field
func EventToDTO(e models.Event) models.EventDTO {
	// ERROR: missing Handler field
	return models.EventDTO{
		Name: e.Name,
	}
}

// Similarly, in strict mode, Notify must be handled
func MessageToDTO(m models.Message) models.MessageDTO {
	// ERROR: missing Notify field
	return models.MessageDTO{
		ID:      m.ID,
		Content: m.Content,
	}
}
