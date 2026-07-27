package sample_indexed_access_clean

import (
	models "converters/20-indexed-access/models"
)

// ConvertCueToDTO converts a single Cue. It is the element converter the delegating
// variants below hand each element to.
func ConvertCueToDTO(c models.Cue) models.CueDTO {
	return models.CueDTO{
		Start: c.Start,
		Text:  c.Text,
	}
}

// ConvertCuesToDTO reads every field through an index expression on the parameter,
// the idiomatic form for a pre-allocated result.
func ConvertCuesToDTO(in []models.Cue) []models.CueDTO {
	out := make([]models.CueDTO, len(in))
	for i := range in {
		out[i] = models.CueDTO{
			Start: in[i].Start,
			Text:  in[i].Text,
		}
	}
	return out
}

// ConvertCuesToDTOClassicLoop is the same conversion driven by a C-style loop,
// which has no range value at all.
func ConvertCuesToDTOClassicLoop(in []models.Cue) []models.CueDTO {
	out := make([]models.CueDTO, len(in))
	for i := 0; i < len(in); i++ {
		out[i] = models.CueDTO{
			Start: in[i].Start,
			Text:  in[i].Text,
		}
	}
	return out
}

// ConvertCuesToDTOMixed reads one field off the range value and the other off the
// index: both spellings in one body.
func ConvertCuesToDTOMixed(in []models.Cue) []models.CueDTO {
	out := make([]models.CueDTO, len(in))
	for i, v := range in {
		out[i] = models.CueDTO{
			Start: v.Start,
			Text:  in[i].Text,
		}
	}
	return out
}

// ConvertCuesToDTOIndexedWrite fills an unnamed result through the index instead of
// building a composite literal per element.
func ConvertCuesToDTOIndexedWrite(in []models.Cue) []models.CueDTO {
	out := make([]models.CueDTO, len(in))
	for i := range in {
		out[i].Start = in[i].Start
		out[i].Text = in[i].Text
	}
	return out
}

// ConvertCuesToDTONamedResult does the same with a named result.
func ConvertCuesToDTONamedResult(in []models.Cue) (out []models.CueDTO) {
	out = make([]models.CueDTO, len(in))
	for i := range in {
		out[i].Start = in[i].Start
		out[i].Text = in[i].Text
	}
	return out
}

// ConvertCuesToDTODelegated delegates each element to ConvertCueToDTO, reaching the
// element by index rather than through a range value.
func ConvertCuesToDTODelegated(in []models.Cue) []models.CueDTO {
	out := make([]models.CueDTO, len(in))
	for i := range in {
		out[i] = ConvertCueToDTO(in[i])
	}
	return out
}

// ConvertCueMapToDTO converts a map through the range value.
func ConvertCueMapToDTO(in map[string]models.Cue) map[string]models.CueDTO {
	out := make(map[string]models.CueDTO, len(in))
	for k, v := range in {
		out[k] = models.CueDTO{
			Start: v.Start,
			Text:  v.Text,
		}
	}
	return out
}

// ConvertCueMapToDTOIndexed converts a map by looking each key back up.
func ConvertCueMapToDTOIndexed(in map[string]models.Cue) map[string]models.CueDTO {
	out := make(map[string]models.CueDTO, len(in))
	for k := range in {
		out[k] = models.CueDTO{
			Start: in[k].Start,
			Text:  in[k].Text,
		}
	}
	return out
}
