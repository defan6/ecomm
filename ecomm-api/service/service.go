package service

import (
	"context"
	productDto "ecomm/ecomm-api/handler/dto/product"
	"ecomm/ecomm-api/storer"
	"ecomm/mapper"
)

type Service struct {
	storer *storer.PostgresStorer
}

func NewService(storer *storer.PostgresStorer) *Service {
	return &Service{storer: storer}
}

func (s *Service) CreateProduct(ctx context.Context, createProductReq *productDto.CreateProductReq) (productDto.ProductRes, error) {
	p := mapper.MapToProductFromCreateProductReq(createProductReq)

	p, err := s.storer.CreateProduct(ctx, p)
	if err != nil {
		return productDto.ProductRes{}, err
	}

	productRes := mapper.MapToProductRes(p)
	return productRes, nil
}

func (s *Service) GetProduct(ctx context.Context, id int64) (productDto.ProductRes, error) {
	p, err := s.storer.GetProduct(ctx, id)
	if err != nil {
		return productDto.ProductRes{}, err
	}
	productRes := mapper.MapToProductRes(p)
	return productRes, nil
}

func (s *Service) GetProducts(ctx context.Context) ([]productDto.ProductRes, error) {
	productList, err := s.storer.GetProducts(ctx)
	if err != nil {
		return []productDto.ProductRes{}, err
	}
	productResList := mapper.MapToProductResList(productList)
	return productResList, nil
}

func (s *Service) UpdateProduct(ctx context.Context, id int64, updateProductReq *productDto.UpdateProductReq) (productDto.ProductRes, error) {
	p := mapper.MapToProductFromUpdateProductReq(updateProductReq)
	p.ID = id
	err := s.storer.UpdateProduct(ctx, p)
	if err != nil {
		return productDto.ProductRes{}, err
	}
	productRes := mapper.MapToProductRes(p)
	return productRes, nil
}

func (s *Service) DeleteProduct(ctx context.Context, id int64) error {
	return s.storer.DeleteProduct(ctx, id)
}
