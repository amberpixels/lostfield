package readmeExample

func ConvertUserToDTO(user User) UserDTO { // want "user\\.Email\\s+→\\s??"
	return UserDTO{
		ID:       user.ID,
		Username: user.Username,
		// Missing: Email
	}
}
