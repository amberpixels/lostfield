package sample_nested

import (
	"converters/12-nested-fields/models/domain"
	"converters/12-nested-fields/models/dto"
)

// Missing deeply nested field in inline declaration
func ConvertEventToDTO_FullInlineDeclaration_Missed_Deep_Field(event domain.Event) dto.EventDTO {
	return dto.EventDTO{
		ID:    event.ID,
		Title: event.Title,
		User: dto.UserDTO{
			ID:   event.User.ID,
			Name: event.User.Name,
			Role: dto.RoleDTO{
				ID: event.User.Role.ID,
				// MISSING: Name: event.User.Role.Name,
			},
			Group: &dto.GroupDTO{
				ID:   event.User.Group.ID,
				Name: event.User.Group.Name,
			},
		},
		Owner: &dto.UserDTO{
			ID:   event.Owner.ID,
			Name: event.Owner.Name,
			Role: dto.RoleDTO{
				ID:   event.Owner.Role.ID,
				Name: event.Owner.Role.Name,
			},
			Group: &dto.GroupDTO{
				ID:   event.Owner.Group.ID,
				Name: event.Owner.Group.Name,
			},
		},
	}
}

// Missing pointer field in inline declaration
func ConvertEventToDTO_FullInlineDeclaration_Missed_Pointer_Field(event domain.Event) dto.EventDTO {
	return dto.EventDTO{
		ID:    event.ID,
		Title: event.Title,
		User: dto.UserDTO{
			ID:   event.User.ID,
			Name: event.User.Name,
			Role: dto.RoleDTO{
				ID:   event.User.Role.ID,
				Name: event.User.Role.Name,
			},
			// MISSING: Group field
		},
		Owner: &dto.UserDTO{
			ID:   event.Owner.ID,
			Name: event.Owner.Name,
			Role: dto.RoleDTO{
				ID:   event.Owner.Role.ID,
				Name: event.Owner.Role.Name,
			},
			Group: &dto.GroupDTO{
				ID:   event.Owner.Group.ID,
				Name: event.Owner.Group.Name,
			},
		},
	}
}

// Missing field in mixed declaration and dot notation approach
func ConvertEventToDTO_Mixed_Missed_Nested_Field(event domain.Event) (result dto.EventDTO) {
	result = dto.EventDTO{
		ID:    event.ID,
		Title: event.Title,
		User: dto.UserDTO{
			ID:   event.User.ID,
			Name: event.User.Name,
		},
		Owner: &dto.UserDTO{
			ID:   event.Owner.ID,
			Name: event.Owner.Name,
		},
	}

	result.User.Role.ID = event.User.Role.ID
	result.User.Role.Name = event.User.Role.Name
	result.User.Group.ID = event.User.Group.ID
	result.User.Group.Name = event.User.Group.Name

	result.Owner.Role.ID = event.Owner.Role.ID
	result.Owner.Role.Name = event.Owner.Role.Name
	// MISSING: result.Owner.Group assignment

	return
}

// Missing entire first-level nested struct
func ConvertEventToDTO_DotNotation_Missed_First_Level(event domain.Event) (result dto.EventDTO) { // want "incomplete converter"
	result.ID = event.ID
	result.Title = event.Title
	// MISSING: result.User completely

	result.Owner = &dto.UserDTO{
		ID:   event.Owner.ID,
		Name: event.Owner.Name,
	}
	result.Owner.Role.ID = event.Owner.Role.ID
	result.Owner.Role.Name = event.Owner.Role.Name
	result.Owner.Group = &dto.GroupDTO{
		ID:   event.Owner.Group.ID,
		Name: event.Owner.Group.Name,
	}

	return
}
