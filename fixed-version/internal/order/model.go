package order

import "time"

const (
	StatusCreated        = "created"
	PaymentStatusPending = "pending"
	PaymentStatusPaid    = "paid"
)

type Order struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	OrderNumber   string    `json:"order_number" gorm:"uniqueIndex;not null"`
	ProductID     uint      `json:"product_id" gorm:"not null"`
	Quantity      int       `json:"quantity" gorm:"not null"`
	CustomerEmail string    `json:"customer_email" gorm:"index;not null"`
	TotalPrice    int       `json:"total_price" gorm:"not null"`
	Status        string    `json:"status" gorm:"not null;default:'created'"`
	PaymentStatus string    `json:"payment_status" gorm:"not null;default:'pending'"`
	PaidAmount    int       `json:"paid_amount" gorm:"not null;default:0"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type CreateOrderRequest struct {
	ProductID     uint   `json:"product_id"`
	Quantity      int    `json:"quantity"`
	CustomerEmail string `json:"customer_email"`
	OrderNumber   string `json:"order_number"`
	SimulateSlow  bool   `json:"simulate_slow"`
}

type DiscountRequest struct {
	Price           int `json:"price"`
	DiscountPercent int `json:"discount_percent"`
}

type DiscountResponse struct {
	Price           int `json:"price"`
	DiscountPercent int `json:"discount_percent"`
	DiscountAmount  int `json:"discount_amount"`
	FinalPrice      int `json:"final_price"`
}
