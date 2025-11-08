package both_or_nothing

import "converters/10-non-marshallable-fields/models"

// In "both-or-nothing" mode, the Handler field only matters if it exists in BOTH Event and EventDTO
// Since EventDTO doesn't have a Handler field, it's ignored here
func EventToDTO(e models.Event) models.EventDTO {
	return models.EventDTO{
		Name: e.Name,
	}
}

// Similarly, Notify only matters if BOTH Message and MessageDTO have it
// Since MessageDTO doesn't have Notify, it's ignored
func MessageToDTO(m models.Message) models.MessageDTO {
	return models.MessageDTO{
		ID:      m.ID,
		Content: m.Content,
	}
}
