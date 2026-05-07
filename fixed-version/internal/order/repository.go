package order

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) WithDB(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, order *Order) error {
	return r.db.WithContext(ctx).Create(order).Error
}

func (r *Repository) FindByID(ctx context.Context, id uint) (*Order, error) {
	var order Order
	err := r.db.WithContext(ctx).First(&order, id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *Repository) FindByIDForUpdate(ctx context.Context, id uint) (*Order, error) {
	var order Order
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).First(&order, id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *Repository) IncrementPaidAmount(ctx context.Context, id uint, amount int, paymentStatus string) error {
	return r.db.WithContext(ctx).
		Model(&Order{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"payment_status": paymentStatus,
			"paid_amount":    gorm.Expr("paid_amount + ?", amount),
		}).Error
}

func (r *Repository) SetPaymentState(ctx context.Context, id uint, paidAmount int, paymentStatus string) error {
	return r.db.WithContext(ctx).
		Model(&Order{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"payment_status": paymentStatus,
			"paid_amount":    paidAmount,
		}).Error
}
