package mapper

import (
	"ecomm/domain"
	productDto "ecomm/ecomm-api/handler/dto/product"
)

func MapToProductRes(product *domain.Product) productDto.ProductRes {
	return productDto.ProductRes{
		ID:           product.ID,
		Name:         product.Name,
		Image:        product.Image,
		Category:     product.Category,
		Description:  product.Description,
		Rating:       product.Rating,
		Price:        product.Price,
		CountInStock: product.CountInStock,
		CreatedAt:    product.CreatedAt,
		UpdatedAt:    product.UpdatedAt,
	}
}

func MapToProductFromCreateProductReq(productReq *productDto.CreateProductReq) *domain.Product {
	return &domain.Product{
		Name:         productReq.Name,
		Image:        productReq.Image,
		Category:     productReq.Category,
		Description:  productReq.Description,
		Rating:       productReq.Rating,
		NumReviews:   productReq.NumReviews,
		Price:        productReq.Price,
		CountInStock: productReq.CountInStock,
	}
}

func MapToProductFromUpdateProductReq(productReq *productDto.UpdateProductReq) *domain.Product {
	return &domain.Product{
		Name:         productReq.Name,
		Image:        productReq.Image,
		Category:     productReq.Category,
		Description:  productReq.Description,
		Rating:       productReq.Rating,
		NumReviews:   productReq.NumReviews,
		Price:        productReq.Price,
		CountInStock: productReq.CountInStock,
	}
}

func MapToProductResList(products []*domain.Product) []productDto.ProductRes {
	productResList := make([]productDto.ProductRes, 0)

	for _, product := range products {
		productRes := MapToProductRes(product)
		productResList = append(productResList, productRes)
	}

	return productResList
}
