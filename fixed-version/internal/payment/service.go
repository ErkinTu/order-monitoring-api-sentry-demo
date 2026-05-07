package payment

import (
	"context"
	"errors"
	"fmt"
	"time"

	ordermodule "github.com/example/order-monitoring-api-fixed/internal/order"

	"gorm.io/gorm"
)

type AppError struct {
	Status  int
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err == nil {
		return e.Message
	}
	return e.Message + ": " + e.Err.Error()
}

type Service struct {
	db       *gorm.DB
	payments *Repository
	orders   *ordermodule.Repository
}

func NewService(db *gorm.DB, payments *Repository, orders *ordermodule.Repository) *Service {
	return &Service{
		db:       db,
		payments: payments,
		orders:   orders,
	}
}

func (s *Service) Capture(ctx context.Context, orderID uint) (*Payment, error) {
	var payment *Payment

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		orderRepo := s.orders.WithDB(tx)
		paymentRepo := s.payments.WithDB(tx)

		order, err := orderRepo.FindByIDForUpdate(ctx, orderID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return &AppError{Status: 404, Message: "order not found", Err: err}
			}
			return err
		}

		if order.PaymentStatus == ordermodule.PaymentStatusPaid {
			return &AppError{Status: 409, Message: "order already paid"}
		}

		payment = &Payment{
			OrderID:           order.ID,
			Amount:            order.TotalPrice,
			Status:            StatusCaptured,
			ProviderReference: generateProviderReference(),
		}

		if err := paymentRepo.Create(ctx, payment); err != nil {
			return err
		}

		return orderRepo.SetPaymentState(ctx, order.ID, order.PaidAmount+payment.Amount, ordermodule.PaymentStatusPaid)
	})
	if err != nil {
		return nil, err
	}

	return payment, nil
}

func (s *Service) ListByOrderID(ctx context.Context, orderID uint) ([]Payment, error) {
	if _, err := s.orders.FindByID(ctx, orderID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &AppError{Status: 404, Message: "order not found", Err: err}
		}
		return nil, err
	}

	return s.payments.ListByOrderID(ctx, orderID)
}

func generateProviderReference() string {
	return fmt.Sprintf("pay_%d", time.Now().UnixNano())
}
