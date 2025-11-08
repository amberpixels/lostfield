package models

// User represents a domain user with both public and private fields
type User struct {
	ID       string // Public field
	Name     string // Public field
	id       string // Private field
	password string // Private field
}

// UserDTO is the data transfer object with only public fields
type UserDTO struct {
	ID   string
	Name string
}
