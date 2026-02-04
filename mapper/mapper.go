package mapper

import (
	"ecomm/domain"
	authDto "ecomm/ecomm-api/handler/dto/auth"
	orderDto "ecomm/ecomm-api/handler/dto/order"
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

func MapToOrderFromCreateOrderReq(orderReq *orderDto.CreateOrderReq) *domain.Order {
	var orderItems []domain.OrderItem

	for _, item := range orderReq.Items {
		orderItem := mapToOrderItemFromOrderItemReq(item)
		orderItems = append(orderItems, orderItem)
	}

	return &domain.Order{
		PaymentMethod: orderReq.PaymentMethod,
		Items:         orderItems,
	}
}

func mapToOrderItemFromOrderItemReq(orderItemReq orderDto.CreateOrderItemReq) domain.OrderItem {
	return domain.OrderItem{
		ProductID: orderItemReq.ProductID,
		Quantity:  orderItemReq.Quantity,
	}
}

func mapToOrderItemResFromOrderItem(orderItem domain.OrderItem) orderDto.OrderItemRes {
	return orderDto.OrderItemRes{
		ID:        orderItem.ID,
		Name:      orderItem.Name,
		Quantity:  orderItem.Quantity,
		Image:     orderItem.Image,
		Price:     orderItem.Price,
		ProductID: orderItem.ProductID,
		OrderID:   orderItem.OrderID,
	}
}

func MapToOrderRes(order *domain.Order) orderDto.OrderRes {
	var orderItemsRes []orderDto.OrderItemRes

	for _, item := range order.Items {
		orderItemRes := mapToOrderItemResFromOrderItem(item)
		orderItemsRes = append(orderItemsRes, orderItemRes)
	}

	return orderDto.OrderRes{
		ID:            order.ID,
		PaymentMethod: order.PaymentMethod,
		TaxPrice:      order.TaxPrice,
		ShippingPrice: order.ShippingPrice,
		TotalPrice:    order.TotalPrice,
		Items:         orderItemsRes,
		CreatedAt:     order.CreatedAt,
		UpdatedAt:     order.UpdatedAt,
	}
}

func MapToUserFromRegisterReq(req *authDto.RegisterRequest) *domain.User {
	return &domain.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	}
}

func MapToUserResFromUser(user *domain.User) authDto.RegisterResponse {
	return authDto.RegisterResponse{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}
