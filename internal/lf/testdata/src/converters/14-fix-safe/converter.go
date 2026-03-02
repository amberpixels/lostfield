package fix_safe

import (
	"converters/14-fix-safe/models/domain"
	"converters/14-fix-safe/models/dto"
)

// ConvertUserToDTO_MissingFields is missing Email and Phone fields
func ConvertUserToDTO_MissingFields(user domain.User) (result *dto.UserDTO) { // want "incomplete converter"
	result = &dto.UserDTO{
		ID:   user.ID,
		Name: user.Name,
	}
	return
}
