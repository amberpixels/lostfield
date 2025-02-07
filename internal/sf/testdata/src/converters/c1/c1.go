package c1

import (
	"converters/dbmodel"
	"converters/model"
)

func ConvertSampleToDB(sample model.Sample) *dbmodel.Sample {
	return &dbmodel.Sample{
		ID:       sample.ID,
		Label:    sample.Label,
		Price:    sample.Price,
		Currency: sample.Currency,
	}
}
