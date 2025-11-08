package sample_chained_return

import "converters/7-chained-return/models"

// InputToOutputDirect converts InputModel to OutputModel with direct composite literal return
// This should work fine
func InputToOutputDirect(input *models.InputModel) *models.OutputModel {
	if input == nil {
		return nil
	}

	return &models.OutputModel{
		Name:  input.Name,
		Value: input.Value,
	}
}

// InputToOutputChained converts InputModel to OutputModel with chained method call
// This pattern returns (&Type{fields...}).MethodCall()
// Currently this fails because the linter doesn't detect fields in chained calls
func InputToOutputChained(input *models.InputModel) *models.OutputModel {
	if input == nil {
		return nil
	}

	return (&models.OutputModel{
		Name:  input.Name,
		Value: input.Value,
	}).Clone() // Simulate a method that returns the same type
}

// ProtoVenueConfigToModel converts proto to model with chained Prepare call
// This should detect ALL fields in the composite literal and NOT flag as missing
func ProtoVenueConfigToModel(config *models.VenueConfig) *models.VenueModel {
	if config == nil {
		return nil
	}

	return (&models.VenueModel{
		ID:           config.ID,
		Name:         config.Name,
		CreatedAt:    "now",
		UpdatedAt:    "now",
		IsDeprecated: config.Deprecated,
		MapSlug:      "",
		Priority:     0,
	}).Prepare()
}

// ProtoVenueConfigToModelChainedNoNil - same converter but without nil check
// Testing if nil check affects detection
func ProtoVenueConfigToModelChainedNoNil(config *models.VenueConfig) *models.VenueModel {
	return (&models.VenueModel{
		ID:           config.ID,
		Name:         config.Name,
		CreatedAt:    "now",
		UpdatedAt:    "now",
		IsDeprecated: config.Deprecated,
		MapSlug:      "",
		Priority:     0,
	}).Prepare()
}

// ProtoVenueConfigToModelMissingFields converts proto to model with chained Prepare call
// but intentionally doesn't set all output fields
// This SHOULD flag as missing some output fields
func ProtoVenueConfigToModelMissingFields(config *models.VenueConfig) *models.VenueModel {
	if config == nil {
		return nil
	}

	return (&models.VenueModel{
		ID:   config.ID,
		Name: config.Name,
		// Missing: CreatedAt, UpdatedAt, IsDeprecated, MapSlug, Priority
	}).Prepare()
}
