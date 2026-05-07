package order

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	productrepo "github.com/example/order-monitoring-api-fixed/internal/product"

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
	orders   *Repository
	products *productrepo.Repository
}

func NewService(orders *Repository, products *productrepo.Repository) *Service {
	return &Service{orders: orders, products: products}
}

func (s *Service) Create(ctx context.Context, req CreateOrderRequest) (*Order, error) {
	if req.ProductID == 0 {
		return nil, &AppError{Status: 400, Message: "product_id is required"}
	}
	if req.Quantity <= 0 {
		return nil, &AppError{Status: 400, Message: "quantity must be greater than zero"}
	}
	if !strings.Contains(req.CustomerEmail, "@") {
		return nil, &AppError{Status: 400, Message: "customer_email must be valid"}
	}
	if req.SimulateSlow {
		return nil, &AppError{Status: 400, Message: "simulate_slow is disabled in fixed version"}
	}

	p, err := s.products.FindByID(ctx, req.ProductID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &AppError{Status: 404, Message: "product not found"}
		}
		return nil, err
	}

	orderNumber := strings.TrimSpace(req.OrderNumber)
	if orderNumber == "" {
		orderNumber = generateOrderNumber()
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
		if isDuplicateKey(err) {
			return nil, &AppError{Status: 409, Message: "order_number already exists", Err: err}
		}
		return nil, err
	}

	return order, nil
}

func (s *Service) CalculateDiscount(req DiscountRequest) (DiscountResponse, error) {
	if req.Price <= 0 {
		return DiscountResponse{}, &AppError{Status: 400, Message: "price must be greater than zero"}
	}
	if req.DiscountPercent <= 0 || req.DiscountPercent > 100 {
		return DiscountResponse{}, &AppError{Status: 400, Message: "discount_percent must be between 1 and 100"}
	}

	discountAmount := req.Price * req.DiscountPercent / 100
	finalPrice := req.Price - discountAmount

	return DiscountResponse{
		Price:           req.Price,
		DiscountPercent: req.DiscountPercent,
		DiscountAmount:  discountAmount,
		FinalPrice:      finalPrice,
	}, nil
}

func generateOrderNumber() string {
	return fmt.Sprintf("ORD-%d", time.Now().UnixNano())
}

func isDuplicateKey(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "duplicate key") || strings.Contains(message, "unique constraint")
}
