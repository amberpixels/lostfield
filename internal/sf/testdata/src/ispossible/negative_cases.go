package ispossible

// Negative test cases - functions that should NOT be detected as converters

// NoParams has no parameters
func NoParams() UserDTO {
	return UserDTO{}
}

// NoResults has no return values
func NoResults(user User) {
	_ = user
}

// NoStructParams only has primitive types
func NoStructParams(id int64, name string) (int64, string) {
	return id, name
}

// UnrelatedTypes has struct params but no naming similarity
func UnrelatedTypes(user User) CompletelyDifferent {
	return CompletelyDifferent{
		Field: user.Username,
	}
}

// SameTypeInOut has the same type as input and output (not a conversion)
func SameTypeInOut(user User) User {
	return user
}

// MultipleUnrelatedStructs has multiple struct params but no clear conversion pair
func MultipleUnrelatedStructs(cat Category, order Order) CompletelyDifferent {
	return CompletelyDifferent{}
}

// OnlyPrimitiveReturn returns only primitives despite struct input
func OnlyPrimitiveReturn(user User) (string, error) {
	return user.Email, nil
}

// OnlyErrorReturn common pattern that should not be a converter
func OnlyErrorReturn(user User) error {
	return nil
}

// SliceToNonSlice incompatible container types (slice input, non-slice output)
func SliceToNonSlice(users []User) UserDTO {
	if len(users) > 0 {
		return ConvertUserToDTO(users[0])
	}
	return UserDTO{}
}

// NonSliceToSlice incompatible container types (non-slice input, slice output)
func NonSliceToSlice(user User) []UserDTO {
	return []UserDTO{ConvertUserToDTO(user)}
}

// MapToSlice incompatible container types (map input, slice output)
func MapToSlice(users map[string]User) []UserDTO {
	result := make([]UserDTO, 0)
	for _, u := range users {
		result = append(result, ConvertUserToDTO(u))
	}
	return result
}

// SliceToMap incompatible container types (slice input, map output)
func SliceToMap(users []User) map[string]UserDTO {
	result := make(map[string]UserDTO)
	for _, u := range users {
		result[u.Username] = ConvertUserToDTO(u)
	}
	return result
}

// HelperFunction utility function with similar types but different purpose
func HelperFunction(user User, category Category) bool {
	return user.ID > 0
}

// WithContextAndError common signature that should not be a converter
func WithContextAndError(user User) (UserDTO, error) {
	// Even though this looks like a converter, it returns error
	// which is a common pattern for non-converter functions
	return ConvertUserToDTO(user), nil
}
