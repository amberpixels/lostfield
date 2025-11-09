package domain

// Domain Models Relationships
//
// Event
//  ├─ User (value type)
//  │   ├─ Role (value type)
//  │   └─ Group (pointer type)
//  └─ Owner (pointer type of User)
//      ├─ Role (value type)
//      └─ Group (pointer type)

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
