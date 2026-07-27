package sample_forwarding_dirty

import (
	models "converters/21-forwarding/models"
)

// ConvertLocation is a complete converter, mapped in full.
func ConvertLocation(in models.Location) models.LocationDTO {
	return models.LocationDTO{
		ID:    in.ID,
		Title: in.Title,
	}
}

// ConvertLocationMixed_MissingTitle forwards on one branch but builds the output itself
// on another, dropping Title there. Forwarding somewhere does not excuse the branch that
// maps by hand, so this is still reported.
func ConvertLocationMixed_MissingTitle(in models.Location, curator bool) models.LocationDTO { // want "ConvertLocationMixed_MissingTitle"
	if curator {
		return ConvertLocation(in)
	}
	return models.LocationDTO{
		ID: in.ID,
		// Title is missing
	}
}
