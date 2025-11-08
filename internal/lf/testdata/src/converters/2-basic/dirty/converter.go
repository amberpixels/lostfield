package sample_basic

import (
	"converters/2-basic/models/dbmodel"
	"converters/2-basic/models/domain"
)

// ConvertSampleToDB_MissingPrice is missing the Price field
func ConvertSampleToDB_MissingPrice(sample domain.Sample) (result *dbmodel.Sample) { // want "incomplete converter"
	_ = sample.Label
	_ = sample.ID

	result = &dbmodel.Sample{
		Label:    "const label",
		Currency: sample.Currency,
	}
	// Missing: result.Price = sample.Price

	_ = result.ID

	return
}

// ConvertSampleToDB_MissingCurrency is missing the Currency field
func ConvertSampleToDB_MissingCurrency(sample domain.Sample) (result *dbmodel.Sample) { // want "incomplete converter"
	_ = sample.Label
	_ = sample.ID

	result = &dbmodel.Sample{
		Label: "const label",
	}
	result.Price = sample.Price

	_ = result.ID

	return
}

// ConvertSampleToDB_MissingBoth is missing both Price and Currency fields
func ConvertSampleToDB_MissingBoth(sample domain.Sample) (result *dbmodel.Sample) { // want "incomplete converter"
	_ = sample.Label
	_ = sample.ID

	result = &dbmodel.Sample{
		Label: "const label",
	}
	// Missing: result.Price and Currency

	_ = result.ID

	return
}
