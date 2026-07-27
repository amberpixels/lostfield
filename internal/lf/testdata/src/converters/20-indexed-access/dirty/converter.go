package sample_indexed_access_dirty

import (
	models "converters/20-indexed-access/models"
)

// ConvertCuesToDTO_MissingText drops Text while reading through the index.
func ConvertCuesToDTO_MissingText(in []models.Cue) []models.CueDTO { // want "ConvertCuesToDTO_MissingText"
	out := make([]models.CueDTO, len(in))
	for i := range in {
		out[i] = models.CueDTO{
			Start: in[i].Start,
			// Text is missing
		}
	}
	return out
}

// ConvertCuesToDTOIndexedWrite_MissingText drops Text while writing through the index.
func ConvertCuesToDTOIndexedWrite_MissingText(in []models.Cue) []models.CueDTO { // want "ConvertCuesToDTOIndexedWrite_MissingText"
	out := make([]models.CueDTO, len(in))
	for i := range in {
		out[i].Start = in[i].Start
		// Text is missing
	}
	return out
}

// ConvertCuesToDTONamedResult_MissingText is the same with a named result.
func ConvertCuesToDTONamedResult_MissingText(in []models.Cue) (out []models.CueDTO) { // want "ConvertCuesToDTONamedResult_MissingText"
	out = make([]models.CueDTO, len(in))
	for i := range in {
		out[i].Start = in[i].Start
		// Text is missing
	}
	return out
}

// ConvertCueMapToDTO_MissingText drops Text in a map converter.
func ConvertCueMapToDTO_MissingText(in map[string]models.Cue) map[string]models.CueDTO { // want "ConvertCueMapToDTO_MissingText"
	out := make(map[string]models.CueDTO, len(in))
	for k, v := range in {
		out[k] = models.CueDTO{
			Start: v.Start,
			// Text is missing
		}
	}
	return out
}
