package fix_smart

import (
	"converters/15-fix-smart/models/domain"
	"converters/15-fix-smart/models/dto"
)

// ConvertProductToDTO_MissingFields is missing Price and Category
func ConvertProductToDTO_MissingFields(product domain.Product) (result *dto.ProductDTO) { // want "incomplete converter"
	result = &dto.ProductDTO{
		ID:   product.ID,
		Name: product.Name,
	}
	return
}
