package sample_readme_example

import "time"

type User struct {
	ID        int64
	Username  string
	Email     string
	CreatedAt time.Time
}

type UserDTO struct {
	ID       int64
	Username string
	Email    string
}
