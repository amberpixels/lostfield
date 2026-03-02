package sample_slice_inline_dirty

import (
	models "converters/13-slice-inline-mapping/models"
)

// ConvertPetInfosToDTO_MissingWeight converts a slice of PetInfo to PetInfoDTO
// but forgets to map the Weight field in both directions.
func ConvertPetInfosToDTO_MissingWeight(records []models.PetInfo) []models.PetInfoDTO { // want "ConvertPetInfosToDTO_MissingWeight"
	if records == nil {
		return nil
	}
	result := make([]models.PetInfoDTO, len(records))
	for i, rec := range records {
		result[i] = models.PetInfoDTO{
			Name:  rec.Name,
			Breed: rec.Breed,
			Age:   rec.Age,
			// Weight is missing
		}
	}
	return result
}
