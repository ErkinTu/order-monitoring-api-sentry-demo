package order

import (
	"context"
	"time"

	productrepo "github.com/example/order-monitoring-api-broken/internal/product"
)

type Service struct {
	orders   *Repository
	products *productrepo.Repository
}

func NewService(orders *Repository, products *productrepo.Repository) *Service {
	return &Service{orders: orders, products: products}
}

func (s *Service) Create(ctx context.Context, req CreateOrderRequest) (*Order, error) {
	// Intentional performance issue for Sentry tracing demo.
	if req.SimulateSlow {
		time.Sleep(3 * time.Second)
	}

	p, err := s.products.FindByID(ctx, req.ProductID)
	if err != nil {
		return nil, err
	}

	// Intentional bug: if order_number is empty, every order gets the same number.
	// The second request will trigger PostgreSQL duplicate key violation.
	orderNumber := req.OrderNumber
	if orderNumber == "" {
		orderNumber = "ORDER-DEMO-DUPLICATE"
	}

	order := &Order{
		OrderNumber:   orderNumber,
		ProductID:     req.ProductID,
		Quantity:      req.Quantity,
		CustomerEmail: req.CustomerEmail,
		TotalPrice:    p.Price * req.Quantity,
		Status:        StatusCreated,
		PaymentStatus: PaymentStatusPending,
		PaidAmount:    0,
	}

	if err := s.orders.Create(ctx, order); err != nil {
		return nil, err
	}

	return order, nil
}

func (s *Service) CalculateDiscount(req DiscountRequest) DiscountResponse {
	// Intentional bug: when discount_percent = 0, Go panics with integer divide by zero.
	// Correct formula must be: price * discount_percent / 100.
	discountAmount := req.Price / req.DiscountPercent
	finalPrice := req.Price - discountAmount

	return DiscountResponse{
		Price:           req.Price,
		DiscountPercent: req.DiscountPercent,
		DiscountAmount:  discountAmount,
		FinalPrice:      finalPrice,
	}
}
