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
