package chainedReturn

// InputToOutputDirect converts InputModel to OutputModel with direct composite literal return
// This should work fine
func InputToOutputDirect(input *InputModel) *OutputModel {
	if input == nil {
		return nil
	}

	return &OutputModel{
		Name:  input.Name,
		Value: input.Value,
	}
}

// InputToOutputChained converts InputModel to OutputModel with chained method call
// This pattern returns (&Type{fields...}).MethodCall()
// Currently this fails because the linter doesn't detect fields in chained calls
func InputToOutputChained(input *InputModel) *OutputModel {
	if input == nil {
		return nil
	}

	return (&OutputModel{
		Name:  input.Name,
		Value: input.Value,
	}).Clone() // Simulate a method that returns the same type
}

// Clone is a helper method
func (o *OutputModel) Clone() *OutputModel {
	if o == nil {
		return nil
	}
	return &OutputModel{
		Name:  o.Name,
		Value: o.Value,
	}
}

// ProtoVenueConfigToModel converts proto to model with chained Prepare call
// This should detect ALL fields in the composite literal and NOT flag as missing
func ProtoVenueConfigToModel(config *VenueConfig) *VenueModel {
	if config == nil {
		return nil
	}

	return (&VenueModel{
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
func ProtoVenueConfigToModelChainedNoNil(config *VenueConfig) *VenueModel {
	return (&VenueModel{
		ID:           config.ID,
		Name:         config.Name,
		CreatedAt:    "now",
		UpdatedAt:    "now",
		IsDeprecated: config.Deprecated,
		MapSlug:      "",
		Priority:     0,
	}).Prepare()
}

// Prepare is a helper method
func (vm *VenueModel) Prepare() *VenueModel {
	return vm
}

// ProtoVenueConfigToModelMissingFields converts proto to model with chained Prepare call
// but intentionally doesn't set all output fields
// This SHOULD flag as missing some output fields
func ProtoVenueConfigToModelMissingFields(config *VenueConfig) *VenueModel {
	if config == nil {
		return nil
	}

	return (&VenueModel{
		ID:      config.ID,
		Name:    config.Name,
		// Missing: CreatedAt, UpdatedAt, IsDeprecated, MapSlug, Priority
	}).Prepare()
}

// TestChainedWithMethods uses method calls instead of field accesses
// This tests if we properly detect composite literals when input uses method calls
func (vc *VenueConfig) GetID() int {
	return vc.ID
}

func (vc *VenueConfig) GetName() string {
	return vc.Name
}

func (vc *VenueConfig) IsDeprecated() bool {
	return vc.Deprecated
}
