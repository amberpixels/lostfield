package sample_nested

import (
	"converters/12-nested-fields/models/domain"
	"converters/12-nested-fields/models/dto"
)

// Pure: All fields are naturally used in assignments, no blank identifier fixes needed
func ConvertEventToDTO_Pure_FullInlineDeclaration(event domain.Event) dto.EventDTO {
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

// Pure: All fields naturally used via dot notation assignments
func ConvertEventToDTO_Pure_DotNotation(event domain.Event) (result dto.EventDTO) {
	result.ID = event.ID
	result.Title = event.Title
	result.User.ID = event.User.ID
	result.User.Name = event.User.Name
	result.User.Role.ID = event.User.Role.ID
	result.User.Role.Name = event.User.Role.Name
	result.User.Group.ID = event.User.Group.ID
	result.User.Group.Name = event.User.Group.Name
	result.Owner = &dto.UserDTO{}
	result.Owner.ID = event.Owner.ID
	result.Owner.Name = event.Owner.Name
	result.Owner.Role.ID = event.Owner.Role.ID
	result.Owner.Role.Name = event.Owner.Role.Name
	result.Owner.Group = &dto.GroupDTO{}
	result.Owner.Group.ID = event.Owner.Group.ID
	result.Owner.Group.Name = event.Owner.Group.Name
	return
}

// Fixed: Most fields properly assigned, but some need blank identifier fixes
func ConvertEventToDTO_Fixed_MixedDeclaration(event domain.Event) (result dto.EventDTO) {
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
	result.Owner.Group.ID = event.Owner.Group.ID
	result.Owner.Group.Name = event.Owner.Group.Name

	return
}

// FullyFixed: All fields are only acknowledged via blank identifiers, not properly linked
func ConvertEventToDTO_FullyFixed_BlankIdentifiers(event domain.Event) (result dto.EventDTO) {
	_ = event.ID
	_ = event.Title
	_ = event.User.ID
	_ = event.User.Name
	_ = event.User.Role.ID
	_ = event.User.Role.Name
	_ = event.User.Group.ID
	_ = event.User.Group.Name
	_ = event.Owner.ID
	_ = event.Owner.Name
	_ = event.Owner.Role.ID
	_ = event.Owner.Role.Name
	_ = event.Owner.Group.ID
	_ = event.Owner.Group.Name

	_ = result.ID
	_ = result.Title
	_ = result.User.ID
	_ = result.User.Name
	_ = result.User.Role.ID
	_ = result.User.Role.Name
	_ = result.User.Group.ID
	_ = result.User.Group.Name
	_ = result.Owner.ID
	_ = result.Owner.Name
	_ = result.Owner.Role.ID
	_ = result.Owner.Role.Name
	_ = result.Owner.Group.ID
	_ = result.Owner.Group.Name

	return
}
