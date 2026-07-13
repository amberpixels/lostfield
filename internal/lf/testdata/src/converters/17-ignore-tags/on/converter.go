package sample_ignore_tags_on

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

// ConvertProduct skips Secret (tagged lostfield:"ignore"), Audit (matched by
// the bare-key entry "internal") and Comments (tagged json:"-").
// This package is exercised with ignore-tags configured for all three entries,
// so no diagnostics are expected.
func ConvertProduct(p Product) ProductDTO {
	return ProductDTO{
		ID:   p.ID,
		Name: p.Name,
	}
}
