package domain

type Role struct {
	ID   string
	Name string
}

type Group struct {
	ID   string
	Name string
}

type User struct {
	ID    string
	Name  string
	Role  Role
	Group *Group
}

type Event struct {
	ID    string
	Title string
	User  User
	Owner *User
}
