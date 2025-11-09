package dto

type RoleDTO struct {
	ID   string
	Name string
}

type GroupDTO struct {
	ID   string
	Name string
}

type UserDTO struct {
	ID    string
	Name  string
	Role  RoleDTO
	Group *GroupDTO
}

type EventDTO struct {
	ID    string
	Title string
	User  UserDTO
	Owner *UserDTO
}
