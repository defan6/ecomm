package service

import (
	"context"
	"ecomm/domain"
	"ecomm/ecomm-api/handler/dto/order"
	"ecomm/mapper"
	"errors"
	"fmt"
)

type OrderService struct {
	orderStorer OrderStorer
}

func NewOrderService(orderStorer OrderStorer) *OrderService {
	return &OrderService{orderStorer: orderStorer}
}

type OrderStorer interface {
	CreateOrder(ctx context.Context, order *domain.Order) (*domain.Order, error)
	GetProductsByIDs(ctx context.Context, ids []int64) ([]*domain.Product, error)
}

func (s *OrderService) CreateOrder(ctx context.Context, createOrderReq *orderDto.CreateOrderReq) (orderDto.OrderRes, error) {
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

	products, err := s.orderStorer.GetProductsByIDs(ctx, productIDs)
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

	createdOrder, err := s.orderStorer.CreateOrder(ctx, &orderToCreate)
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
