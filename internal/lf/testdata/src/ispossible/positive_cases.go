package ispossible

// Positive test cases - functions that SHOULD be detected as converters

// ConvertUserToDTO converts User to UserDTO (basic struct-to-struct)
func ConvertUserToDTO(user User) UserDTO {
	return UserDTO{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		FullName: user.FirstName + " " + user.LastName,
	}
}

// ConvertUserPtrToDTO converts pointer to User to UserDTO
func ConvertUserPtrToDTO(user *User) UserDTO {
	return UserDTO{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
	}
}

// ConvertUserToDTOPtr converts User to pointer to UserDTO
func ConvertUserToDTOPtr(user User) *UserDTO {
	return &UserDTO{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
	}
}

// ConvertUserPtrToDTOPtr converts pointer to User to pointer to UserDTO
func ConvertUserPtrToDTOPtr(user *User) *UserDTO {
	return &UserDTO{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
	}
}

// ConvertUsersToDTO converts slice of Users to slice of UserDTOs
func ConvertUsersToDTO(users []User) []UserDTO {
	result := make([]UserDTO, len(users))
	for i, user := range users {
		result[i] = ConvertUserToDTO(user)
	}
	return result
}

// ConvertUserSlicePtrToDTO converts slice of User pointers to slice of UserDTOs
func ConvertUserSlicePtrToDTO(users []*User) []UserDTO {
	result := make([]UserDTO, len(users))
	for i, user := range users {
		result[i] = ConvertUserPtrToDTO(user)
	}
	return result
}

// ConvertProductToResponse converts Product to ProductResponse (different suffix pattern)
func ConvertProductToResponse(prod Product) ProductResponse {
	return ProductResponse{
		ID:      prod.ID,
		Name:    prod.Name,
		Price:   prod.Price,
		InStock: prod.Stock > 0,
	}
}

// ConvertProductMap converts map of Products to map of ProductResponses
func ConvertProductMap(products map[string]Product) map[string]ProductResponse {
	result := make(map[string]ProductResponse)
	for k, v := range products {
		result[k] = ConvertProductToResponse(v)
	}
	return result
}

// TransformUserToDTO alternative naming convention (Transform instead of Convert)
func TransformUserToDTO(u User) UserDTO {
	return UserDTO{
		ID:       u.ID,
		Username: u.Username,
	}
}

// UserToDTO short naming convention
func UserToDTO(u User) UserDTO {
	return UserDTO{
		ID:       u.ID,
		Username: u.Username,
	}
}

// ToUserDTO even shorter naming convention
func ToUserDTO(u User) UserDTO {
	return UserDTO{
		ID:       u.ID,
		Username: u.Username,
	}
}

// BuildUserDTOFromUser builder-style naming
func BuildUserDTOFromUser(u User) UserDTO {
	return UserDTO{
		ID:       u.ID,
		Username: u.Username,
	}
}

// FromDomain converts domain Apple to database model DbApple with embedded GormModel
// This tests handling of embedded structs - fields from GormModel (ID, CreatedAt, UpdatedAt)
// are accessed directly on dbapple and should be recognized as part of the embedding.
// This matches the real-world pattern where you manually set each field of embedded struct.
func FromDomain(a Apple) DbApple {
	dbapple := DbApple{}
	dbapple.ID = a.ID
	dbapple.CreatedAt = 1234567890
	dbapple.UpdatedAt = 1234567890
	dbapple.Kind = a.Kind
	dbapple.Color = a.Color
	return dbapple
}
