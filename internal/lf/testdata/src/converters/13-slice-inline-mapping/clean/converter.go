package sample_slice_inline_clean

import (
	models "converters/13-slice-inline-mapping/models"
)

// ConvertPetInfosToDTO converts a slice of PetInfo to a slice of PetInfoDTO
// using inline composite literal mapping (not delegating to another function).
// This should NOT produce any diagnostics because all fields are properly mapped.
func ConvertPetInfosToDTO(records []models.PetInfo) []models.PetInfoDTO {
	if records == nil {
		return nil
	}
	result := make([]models.PetInfoDTO, len(records))
	for i, rec := range records {
		result[i] = models.PetInfoDTO{
			Name:   rec.Name,
			Breed:  rec.Breed,
			Age:    rec.Age,
			Weight: rec.Weight,
		}
	}
	return result
}

// ConvertPetInfosToDTOAppend same as above but using append pattern.
func ConvertPetInfosToDTOAppend(records []models.PetInfo) []models.PetInfoDTO {
	if records == nil {
		return nil
	}
	var result []models.PetInfoDTO
	for _, rec := range records {
		result = append(result, models.PetInfoDTO{
			Name:   rec.Name,
			Breed:  rec.Breed,
			Age:    rec.Age,
			Weight: rec.Weight,
		})
	}
	return result
}
