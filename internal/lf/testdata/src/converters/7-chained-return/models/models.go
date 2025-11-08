package models

// InputModel represents the input
type InputModel struct {
	Name  string
	Value int
}

// OutputModel represents the output
type OutputModel struct {
	Name  string
	Value int
}

// Clone is a helper method
func (o *OutputModel) Clone() *OutputModel {
	if o == nil {
		return nil
	}
	return &OutputModel{
		Name:  o.Name,
		Value: o.Value,
	}
}

// Prepare is a helper method
func (vm *VenueModel) Prepare() *VenueModel {
	return vm
}

// VenueConfig represents input venue configuration
type VenueConfig struct {
	ID         int
	Name       string
	Deprecated bool
}

// TestChainedWithMethods uses method calls instead of field accesses
// This tests if we properly detect composite literals when input uses method calls
func (vc *VenueConfig) GetID() int {
	return vc.ID
}

func (vc *VenueConfig) GetName() string {
	return vc.Name
}

func (vc *VenueConfig) IsDeprecated() bool {
	return vc.Deprecated
}

// VenueModel represents output venue model
type VenueModel struct {
	ID           int
	Name         string
	CreatedAt    string
	UpdatedAt    string
	IsDeprecated bool
	MapSlug      string
	Priority     int
}
