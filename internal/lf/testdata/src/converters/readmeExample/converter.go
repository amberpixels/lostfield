package readmeExample

func ConvertUserToDTO(user User) UserDTO { // want "detected as converter.*missing fields"
	return UserDTO{
		ID:       user.ID,
		Username: user.Username,
		// Missing: Email
	}
}
