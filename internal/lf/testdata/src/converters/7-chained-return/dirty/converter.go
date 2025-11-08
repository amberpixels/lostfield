package sample_chained_return

import (
	"converters/7-chained-return/models"
)

// InputToOutput has missing fields despite chained return
func InputToOutput(input *models.InputModel) *models.OutputModel { // want "incomplete converter"
	if input == nil {
		return nil
	}

	return (&models.OutputModel{
		Name: input.Name,
		// Missing: Value
	}).Clone()
}

// VenueConfigToModel is missing several output fields
func VenueConfigToModel(config *models.VenueConfig) *models.VenueModel { // want "incomplete converter"
	if config == nil {
		return nil
	}

	return (&models.VenueModel{
		ID:   config.ID,
		Name: config.Name,
		// Missing: CreatedAt, UpdatedAt, IsDeprecated, MapSlug, Priority
	}).Prepare()
}
