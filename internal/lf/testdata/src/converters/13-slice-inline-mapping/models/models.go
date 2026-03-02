package modelsSliceInline

// PetInfo represents a pet in the database layer
type PetInfo struct {
	Name   string
	Breed  string
	Age    int
	Weight float64
}

// PetInfoDTO represents a pet in the API layer
type PetInfoDTO struct {
	Name   string
	Breed  string
	Age    int
	Weight float64
}
