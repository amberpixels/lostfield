package default_mode

import "converters/11-private-fields/models"

// In default mode (IncludePrivateFields=false), private fields are ignored
// so we only need to handle public fields - this converter passes
func UserToDTO(u models.User) models.UserDTO {
	return models.UserDTO{
		ID:   u.ID,
		Name: u.Name,
		// Private fields (id, password) are NOT checked
	}
}
