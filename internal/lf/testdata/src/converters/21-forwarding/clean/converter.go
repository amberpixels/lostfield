package sample_forwarding_clean

import (
	"errors"

	models "converters/21-forwarding/models"
)

// ConvertLocation is a complete converter, mapped in full.
func ConvertLocation(in models.Location) models.LocationDTO {
	return models.LocationDTO{
		ID:    in.ID,
		Title: in.Title,
	}
}

// ConvertLocationCurated is another complete converter.
func ConvertLocationCurated(in models.Location) models.LocationDTO {
	return models.LocationDTO{
		ID:    in.ID,
		Title: "* " + in.Title,
	}
}

// ConvertLocationForBranch picks one of them and forwards the whole input.
func ConvertLocationForBranch(in models.Location, curator bool) models.LocationDTO {
	if curator {
		return ConvertLocationCurated(in)
	}
	return ConvertLocation(in)
}

// ConvertLocationAlias forwards on a single return.
func ConvertLocationAlias(in models.Location) models.LocationDTO {
	return ConvertLocation(in)
}

// ConvertLocationChecked forwards after an early error return. The zero literal
// carries no field mapping, so it does not count as building the output.
func ConvertLocationChecked(in models.Location, ok bool) (models.LocationDTO, error) {
	if !ok {
		return models.LocationDTO{}, errors.New("not allowed")
	}
	return ConvertLocation(in), nil
}
