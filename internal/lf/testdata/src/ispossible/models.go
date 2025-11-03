package ispossible

// Test models for IsPossibleConverter tests

// User represents a domain user model
type User struct {
	ID        int64
	Username  string
	Email     string
	FirstName string
	LastName  string
}

// UserDTO represents a data transfer object for User
type UserDTO struct {
	ID       int64
	Username string
	Email    string
	FullName string
}

// Product represents a product model
type Product struct {
	ID    string
	Name  string
	Price float64
	Stock int
}

// ProductResponse represents an API response for Product
type ProductResponse struct {
	ID         string
	Name       string
	Price      float64
	InStock    bool
	Categories []string
}

// Category represents a category
type Category struct {
	ID   int
	Name string
}

// Order represents an order
type Order struct {
	ID         string
	TotalPrice float64
	Status     string
}

// UnrelatedType represents a type with no conversion relationship
type UnrelatedType struct {
	Data string
}

// CompletelyDifferent has no naming similarity to any other type
type CompletelyDifferent struct {
	Field string
}

// Decorator represents a decorator with config
type Decorator struct {
	Value string
}

// DecoratorConfig represents configuration for Decorator
type DecoratorConfig struct {
	Setting string
}

// GormModel represents a GORM model with base fields (similar to real gorm.Model)
type GormModel struct {
	ID        int64
	CreatedAt int64
	UpdatedAt int64
}

// DbApple represents a database model with embedded GormModel
type DbApple struct {
	GormModel
	Kind  string
	Color string
}

// Apple represents a domain model
type Apple struct {
	ID    int64
	Kind  string
	Color string
}
