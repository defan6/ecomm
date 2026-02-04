package service

import (
	"context"
	"ecomm/domain"
	orderDto "ecomm/ecomm-api/handler/dto/order"
	productDto "ecomm/ecomm-api/handler/dto/product"
	"ecomm/ecomm-api/storer"
	"ecomm/mapper"
	"errors"
	"fmt"
)

type Service struct {
	storer *storer.PostgresStorer
}

var productNotFoundError *storer.NotFoundError

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
		if errors.As(err, &productNotFoundError) {
			return productDto.ProductRes{}, &ErrNotFound{
				Op:        productNotFoundError.Op,
				ID:        productNotFoundError.ID,
				Resource:  productNotFoundError.Resource,
				Timestamp: productNotFoundError.Timestamp,
				Err:       err,
			}
		}
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
		if errors.As(err, &productNotFoundError) {
			return productDto.ProductRes{}, &ErrNotFound{
				Op:        productNotFoundError.Op,
				ID:        productNotFoundError.ID,
				Resource:  productNotFoundError.Resource,
				Timestamp: productNotFoundError.Timestamp,
				Err:       err,
			}
		}
		return productDto.ProductRes{}, err
	}
	productRes := mapper.MapToProductRes(p)
	return productRes, nil
}

func (s *Service) DeleteProduct(ctx context.Context, id int64) error {
	err := s.storer.DeleteProduct(ctx, id)
	if err != nil {
		if errors.As(err, &productNotFoundError) {
			return &ErrNotFound{
				Op:        productNotFoundError.Op,
				ID:        productNotFoundError.ID,
				Resource:  productNotFoundError.Resource,
				Timestamp: productNotFoundError.Timestamp,
				Err:       err,
			}
		}
		return err
	}
	return nil
}

func (s *Service) CreateOrder(ctx context.Context, createOrderReq *orderDto.CreateOrderReq) (orderDto.OrderRes, error) {
	op := "createOrder"

	if createOrderReq.PaymentMethod == "" {
		return orderDto.OrderRes{}, errors.New("payment method is required")
	}

	if createOrderReq.Items == nil || len(createOrderReq.Items) == 0 {
		return orderDto.OrderRes{}, errors.New("orderItems is empty")
	}

	if err := isValidOrderItems(createOrderReq.Items); err != nil {
		return orderDto.OrderRes{}, err
	}

	productIDs := make([]int64, 0, len(createOrderReq.Items))
	uniqueProductIDs := make(map[int64]struct{})
	for _, item := range createOrderReq.Items {
		if _, exists := uniqueProductIDs[item.ProductID]; !exists {
			uniqueProductIDs[item.ProductID] = struct{}{}
			productIDs = append(productIDs, item.ProductID)
		}
	}

	products, err := s.storer.GetProductsByIDs(ctx, productIDs)
	if err != nil {
		return orderDto.OrderRes{}, fmt.Errorf("failed to get products for order: %w", err)
	}

	if len(products) != len(uniqueProductIDs) {
		return orderDto.OrderRes{}, NewErrNotFoundProductForOrder(op, "product", nil)
	}

	productMap := make(map[int64]*domain.Product, len(products))

	for _, product := range products {
		productMap[product.ID] = product
	}

	domainItems := make([]domain.OrderItem, 0, len(createOrderReq.Items))
	var itemsPrice float64 = 0

	for _, item := range createOrderReq.Items {
		product := productMap[item.ProductID]

		if product.CountInStock < item.Quantity {
			return orderDto.OrderRes{},
				NewNotEnoughStock(op, "product", product.ID, item.Quantity, product.CountInStock, nil)
		}

		orderItem := domain.OrderItem{
			Name:      product.Name,
			Quantity:  item.Quantity,
			Image:     product.Image,
			Price:     product.Price,
			ProductID: product.ID,
		}
		domainItems = append(domainItems, orderItem)

		itemsPrice += product.Price * float64(item.Quantity)
	}

	const taxRate = 0.1
	const shippingPrice = 150

	taxPrice := itemsPrice * taxRate
	totalPrice := itemsPrice + taxPrice + shippingPrice

	orderToCreate := domain.Order{
		PaymentMethod: createOrderReq.PaymentMethod,
		TaxPrice:      taxPrice,
		ShippingPrice: shippingPrice,
		TotalPrice:    totalPrice,
		Items:         domainItems,
	}

	createdOrder, err := s.storer.CreateOrder(ctx, &orderToCreate)
	if err != nil {
		return orderDto.OrderRes{}, fmt.Errorf("failed to create order: %w", err)
	}

	orderRes := mapper.MapToOrderRes(createdOrder)

	return orderRes, nil

}

func isValidOrderItems(items []orderDto.CreateOrderItemReq) error {
	for _, item := range items {
		if item.Quantity <= 0 {
			return errors.New("invalid quantity")
		}
		if item.ProductID < 0 {
			return errors.New("invalid product id")
		}
	}
	return nil
}
