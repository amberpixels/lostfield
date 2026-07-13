package sample_ignore_tags_off

// Product is the input model. Some fields are tagged to be skipped by the linter.
type Product struct {
	ID       int64
	Name     string
	Secret   string `lostfield:"ignore"`
	Audit    string `internal:"true"`
	Comments string `json:"-"`
}

// ProductDTO is the output model.
type ProductDTO struct {
	ID   int64
	Name string
}

// ConvertProduct skips the tagged fields. Without ignore-tags configured,
// the tags have no effect and the skipped fields are reported.
func ConvertProduct(p Product) ProductDTO { // want "incomplete converter with missing fields: p.Secret, p.Audit, p.Comments"
	return ProductDTO{
		ID:   p.ID,
		Name: p.Name,
	}
}
