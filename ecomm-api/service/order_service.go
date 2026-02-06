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
	GetOrder(ctx context.Context, id int64) (*domain.Order, error)
	UpdateOrder(ctx context.Context, updatedOrder *domain.Order) (*domain.Order, error)
	GetProductsByIDs(ctx context.Context, ids []int64) ([]*domain.Product, error)
}

func (s *OrderService) CreateOrder(ctx context.Context, createOrderReq *orderDto.CreateOrderReq) (orderDto.OrderRes, error) {
	op := "createOrder"

	if err := validateCreateOrderReq(createOrderReq); err != nil {
		return orderDto.OrderRes{}, err
	}

	domainItems, itemsPrice, err := s.processOrderRequest(ctx, createOrderReq, op)
	taxPrice, shippingPrice, totalPrice := calculateOrderPrices(itemsPrice)

	orderToCreate := domain.Order{
		PaymentMethod: createOrderReq.PaymentMethod,
		TaxPrice:      taxPrice,
		ShippingPrice: shippingPrice,
		TotalPrice:    totalPrice,
		Items:         domainItems,
		Status:        domain.OrderStatusPending,
		UserID:        createOrderReq.UserID,
	}

	createdOrder, err := s.orderStorer.CreateOrder(ctx, &orderToCreate)
	if err != nil {
		return orderDto.OrderRes{}, fmt.Errorf("failed to create order: %w", err)
	}

	orderRes := mapper.MapToOrderRes(createdOrder)

	return orderRes, nil
}

func (s *OrderService) UpdateOrder(ctx context.Context, id int64, updateOrderRequest *orderDto.UpdateOrderReq) (orderDto.OrderRes, error) {
	op := "updateOrder"

	if err := validateUpdateOrderReq(id, updateOrderRequest); err != nil {
		return orderDto.OrderRes{}, err
	}
	existingOrder, err := s.orderStorer.GetOrder(ctx, id)

	if err != nil {
		return orderDto.OrderRes{}, errors.New("failed to get order")
	}

	orderToUpdate := *existingOrder
	updatePaymentMethodAndStatus(&orderToUpdate, updateOrderRequest.PaymentMethod, updateOrderRequest.Status)
	updatedItems, err := updateOrderItems(updateOrderRequest.Items)
	if err != nil {
		return orderDto.OrderRes{}, fmt.Errorf("failed to update items: %w", err)
	}
	var (
		itemsPrice  float64
		domainItems []domain.OrderItem
	)
	if len(updatedItems) > 0 {
		domainItems, itemsPrice, err = s.processOrderRequest(ctx, &orderDto.CreateOrderReq{Items: updatedItems}, op)
		if err != nil {
			return orderDto.OrderRes{}, err
		}

		orderToUpdate.Items = domainItems
		taxPrice, shippingPrice, totalPrice := calculateOrderPrices(itemsPrice)
		updatePrices(&orderToUpdate, taxPrice, shippingPrice, totalPrice)
	}
	order, err := s.orderStorer.UpdateOrder(ctx, &orderToUpdate)
	if err != nil {
		return orderDto.OrderRes{}, fmt.Errorf("failed to update order: %w", err)
	}
	orderRes := mapper.MapToOrderRes(order)
	return orderRes, nil

}

func updatePrices(orderToUpdate *domain.Order, taxPrice float64, shippingPrice float64, totalPrice float64) {
	orderToUpdate.TaxPrice = taxPrice
	orderToUpdate.ShippingPrice = shippingPrice
	orderToUpdate.TotalPrice = totalPrice
}

func updateOrderItems(items []orderDto.UpdateOrderItemReq) ([]orderDto.CreateOrderItemReq, error) {
	createItems := make([]orderDto.CreateOrderItemReq, len(items))
	if len(items) > 0 {
		for i, item := range items {
			if item.ProductID == nil || item.Quantity == nil {
				return nil, errors.New("product ID and quantity are required for item updates when recalculating prices")
			}
			createItems[i] = orderDto.CreateOrderItemReq{
				ProductID: *item.ProductID,
				Quantity:  *item.Quantity,
			}
		}
	}
	return createItems, nil
}

func updatePaymentMethodAndStatus(orderToUpdate *domain.Order, paymentMethod *string, status *string) {
	if status != nil {
		orderToUpdate.PaymentMethod = *status
	}
	if paymentMethod != nil {
		orderToUpdate.PaymentMethod = *paymentMethod
	}
}

func calculateOrderPrices(itemsPrice float64) (float64, float64, float64) {
	const taxRate = 0.1
	const shippingPrice = 150
	taxPrice := itemsPrice * taxRate
	totalPrice := itemsPrice + taxPrice + shippingPrice
	return taxPrice, shippingPrice, totalPrice
}

func validateCreateOrderReq(req *orderDto.CreateOrderReq) error {
	if req.PaymentMethod == "" {
		return errors.New("payment method is required")
	}

	if req.Items == nil || len(req.Items) == 0 {
		return errors.New("orderItems is empty")
	}

	if err := isValidOrderItems(req.Items); err != nil {
		return err
	}
	return nil
}

func validateUpdateOrderReq(id int64, req *orderDto.UpdateOrderReq) error {
	if id <= 0 {
		return errors.New("order ID is required for update")
	}

	if req.Status != nil {
		if !isValidOrderStatus(*req.Status) {
			return errors.New("invalid order status provided")
		}
	}

	if len(req.Items) > 0 {
		for _, item := range req.Items {
			if item.ProductID == nil || *item.ProductID <= 0 {
				return errors.New("invalid product ID in order item for update")
			}
			if item.Quantity == nil || *item.Quantity <= 0 {
				return errors.New("invalid quantity in order item for update")
			}
		}
	}
	return nil
}

func isValidOrderStatus(status string) bool {
	switch status {
	case domain.OrderStatusPending, domain.OrderStatusProcessing, domain.OrderStatusShipped, domain.OrderStatusDelivered, domain.OrderStatusCancelled:
		return true
	default:
		return false
	}
}

func (s *OrderService) processOrderRequest(ctx context.Context, createOrderReq *orderDto.CreateOrderReq, op string) ([]domain.OrderItem, float64, error) {
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
		return nil, 0, fmt.Errorf("failed to get products for order: %w", err)
	}

	if len(products) != len(uniqueProductIDs) {
		return nil, 0, NewErrNotFoundProductForOrder(op, "product", nil)
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
			return nil, 0, NewNotEnoughStock(op, "product", product.ID, item.Quantity, product.CountInStock, nil)
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

	return domainItems, itemsPrice, nil
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
