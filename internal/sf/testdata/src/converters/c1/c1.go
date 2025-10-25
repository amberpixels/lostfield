package c1

import (
	"converters/dbmodel"
	"converters/model"
)

func ConvertSampleToDB(sample model.Sample) (result *dbmodel.Sample) {
	_ = sample.Label
	_ = sample.ID

	result = &dbmodel.Sample{
		Label:    "const label",
		Currency: sample.Currency,
	}
	result.Price = sample.Price

	_ = result.ID

	return
}
