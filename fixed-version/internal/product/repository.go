package product

import (
	"context"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context) ([]Product, error) {
	var products []Product
	err := r.db.WithContext(ctx).Order("id asc").Find(&products).Error
	return products, err
}

func (r *Repository) FindByID(ctx context.Context, id uint) (*Product, error) {
	var product Product
	err := r.db.WithContext(ctx).First(&product, id).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}
