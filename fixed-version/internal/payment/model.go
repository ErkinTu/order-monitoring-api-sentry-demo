package payment

import "time"

const (
	StatusCaptured = "captured"
)

type Payment struct {
	ID                uint      `json:"id" gorm:"primaryKey"`
	OrderID           uint      `json:"order_id" gorm:"index;not null"`
	Amount            int       `json:"amount" gorm:"not null"`
	Status            string    `json:"status" gorm:"not null;default:'captured'"`
	ProviderReference string    `json:"provider_reference" gorm:"not null"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}
