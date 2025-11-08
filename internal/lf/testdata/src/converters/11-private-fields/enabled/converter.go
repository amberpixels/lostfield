package enabled_mode

import "converters/11-private-fields/models"

// In enabled mode (IncludePrivateFields=true), private fields MUST be handled
// This converter is incomplete - missing password field handling
func UserToDTO(u models.User) models.UserDTO {
	// ERROR: private field 'password' is not read/handled
	// (Note: in real Go code, we'd use u.id and u.password via reflection or unsafe,
	// but this demonstrates the linter's capability to check them)
	return models.UserDTO{
		ID:   u.ID,
		Name: u.Name,
		// Missing: password field
	}
}
