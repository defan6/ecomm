package domain

import "time"

const (
	OrderStatusPending    = "pending"
	OrderStatusProcessing = "processing"
	OrderStatusShipped    = "shipped"
	OrderStatusDelivered  = "delivered"
	OrderStatusCancelled  = "cancelled"
)

type Order struct {
	ID            int64      `db:"id"`
	PaymentMethod string     `db:"payment_method"`
	TaxPrice      float64    `db:"tax_price"`
	ShippingPrice float64    `db:"shipping_price"`
	TotalPrice    float64    `db:"total_price"`
	UserID        int64      `db:"user_id"`
	Status        string     `db:"status"`
	CreatedAt     time.Time  `db:"created_at"`
	UpdatedAt     *time.Time `db:"updated_at"`
	Items         []OrderItem
}
