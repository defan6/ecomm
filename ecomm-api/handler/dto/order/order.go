package orderDto

import "time"

type CreateOrderReq struct {
	PaymentMethod string `json:"payment_method"`
	Items         []CreateOrderItemReq
}

type CreateOrderItemReq struct {
	Quantity  int64 `json:"quantity"`
	ProductID int64 `json:"product_id"`
}

type OrderItemRes struct {
	ID        int64   `json:"id"`
	Name      string  `json:"name"`
	Quantity  int64   `json:"quantity"`
	Image     string  `json:"image"`
	Price     float64 `json:"price"`
	ProductID int64   `json:"product_id"`
	OrderID   int64   `json:"order_id"`
}

type OrderRes struct {
	ID            int64          `json:"id"`
	PaymentMethod string         `json:"payment_method"`
	TaxPrice      float64        `json:"tax_price"`
	ShippingPrice float64        `json:"shipping_price"`
	TotalPrice    float64        `json:"total_price"`
	Items         []OrderItemRes `json:"items"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     *time.Time     `json:"updated_at"`
}
