package ignore

import "converters/10-non-marshallable-fields/models"

// In "ignore" mode, the Handler field (func type) is not checked,
// so we only need to handle the Name field - this converter passes
func EventToDTO(e models.Event) models.EventDTO {
	return models.EventDTO{
		Name: e.Name,
	}
}

// Similarly, the Notify channel field is ignored,
// so only Content and ID need to be handled
func MessageToDTO(m models.Message) models.MessageDTO {
	return models.MessageDTO{
		ID:      m.ID,
		Content: m.Content,
	}
}
