package readmeExample

func ConvertUserToDTO(user User) UserDTO { // want "incomplete converter with missing fields:.*user\\.Email"
	return UserDTO{
		ID:       user.ID,
		Username: user.Username,
		// Missing: Email
	}
}
