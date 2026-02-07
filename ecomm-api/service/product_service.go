package service

import (
	"context"
	"ecomm/domain"
	productDto "ecomm/ecomm-api/handler/dto/product"
	"ecomm/mapper"
	"errors"
)

type ProductService struct {
	productStorer ProductStorer
}
type ProductStorer interface {
	CreateProduct(ctx context.Context, p *domain.Product) (*domain.Product, error)
	GetProduct(ctx context.Context, id int64) (*domain.Product, error)
	GetProducts(ctx context.Context) ([]*domain.Product, error)
	UpdateProduct(ctx context.Context, p *domain.Product) error
	DeleteProduct(ctx context.Context, id int64) error
}

func NewProductService(productStorer ProductStorer) *ProductService {
	return &ProductService{
		productStorer: productStorer,
	}
}

func (s *ProductService) CreateProduct(ctx context.Context, createProductReq *productDto.CreateProductReq) (productDto.ProductRes, error) {
	p := mapper.MapToProductFromCreateProductReq(createProductReq)

	p, err := s.productStorer.CreateProduct(ctx, p)
	if err != nil {
		return productDto.ProductRes{}, err
	}

	productRes := mapper.MapToProductRes(p)
	return productRes, nil
}

func (s *ProductService) GetProduct(ctx context.Context, id int64) (productDto.ProductRes, error) {
	p, err := s.productStorer.GetProduct(ctx, id)
	if err != nil {
		if errors.As(err, &notFoundErr) {
			return productDto.ProductRes{}, NewErrNotFound(notFoundErr.Op, notFoundErr.Resource, id, err)
		}
		return productDto.ProductRes{}, err
	}
	productRes := mapper.MapToProductRes(p)
	return productRes, nil
}

func (s *ProductService) GetProducts(ctx context.Context) ([]productDto.ProductRes, error) {
	productList, err := s.productStorer.GetProducts(ctx)
	if err != nil {
		return []productDto.ProductRes{}, err
	}
	productResList := mapper.MapToProductResList(productList)
	return productResList, nil
}

func (s *ProductService) UpdateProduct(ctx context.Context, id int64, updateProductReq *productDto.UpdateProductReq) (productDto.ProductRes, error) {
	p := mapper.MapToProductFromUpdateProductReq(updateProductReq)
	p.ID = id
	err := s.productStorer.UpdateProduct(ctx, p)
	if err != nil {
		if errors.As(err, &notFoundErr) {
			return productDto.ProductRes{}, &ErrNotFound{
				Op:        notFoundErr.Op,
				ID:        notFoundErr.ID,
				Resource:  notFoundErr.Resource,
				Timestamp: notFoundErr.Timestamp,
				Err:       err,
			}
		}
		return productDto.ProductRes{}, err
	}
	productRes := mapper.MapToProductRes(p)
	return productRes, nil
}

func (s *ProductService) DeleteProduct(ctx context.Context, id int64) error {
	err := s.productStorer.DeleteProduct(ctx, id)
	if err != nil {
		if errors.As(err, &notFoundErr) {
			return &ErrNotFound{
				Op:        notFoundErr.Op,
				ID:        notFoundErr.ID,
				Resource:  notFoundErr.Resource,
				Timestamp: notFoundErr.Timestamp,
				Err:       err,
			}
		}
		return err
	}
	return nil
}
