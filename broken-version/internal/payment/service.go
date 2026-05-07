package payment

import (
	"context"
	"errors"
	"fmt"
	"time"

	ordermodule "github.com/example/order-monitoring-api-broken/internal/order"
)

var ErrOrderAlreadyPaid = errors.New("order already paid")

type Service struct {
	payments *Repository
	orders   *ordermodule.Repository
}

func NewService(payments *Repository, orders *ordermodule.Repository) *Service {
	return &Service{
		payments: payments,
		orders:   orders,
	}
}

func (s *Service) Capture(ctx context.Context, orderID uint) (*Payment, error) {
	order, err := s.orders.FindByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	if order.PaymentStatus == ordermodule.PaymentStatusPaid {
		return nil, ErrOrderAlreadyPaid
	}

	// Intentional race window: concurrent workers can all observe "pending"
	// and proceed to charge the same order before the first write is visible.
	time.Sleep(250 * time.Millisecond)

	payment := &Payment{
		OrderID:           order.ID,
		Amount:            order.TotalPrice,
		Status:            StatusCaptured,
		ProviderReference: generateProviderReference(),
	}

	if err := s.payments.Create(ctx, payment); err != nil {
		return nil, err
	}

	if err := s.orders.IncrementPaidAmount(ctx, order.ID, payment.Amount, ordermodule.PaymentStatusPaid); err != nil {
		return nil, err
	}

	return payment, nil
}

func (s *Service) ListByOrderID(ctx context.Context, orderID uint) ([]Payment, error) {
	if _, err := s.orders.FindByID(ctx, orderID); err != nil {
		return nil, err
	}

	return s.payments.ListByOrderID(ctx, orderID)
}

func generateProviderReference() string {
	return fmt.Sprintf("pay_%d", time.Now().UnixNano())
}
