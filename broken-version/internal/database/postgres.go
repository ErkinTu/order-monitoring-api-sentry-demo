package database

import (
	"log"
	"time"

	"github.com/example/order-monitoring-api-broken/internal/config"
	ordermodel "github.com/example/order-monitoring-api-broken/internal/order"
	productmodel "github.com/example/order-monitoring-api-broken/internal/product"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectWithRetry(cfg config.Config) (*gorm.DB, error) {
	var lastErr error

	for attempt := 1; attempt <= 10; attempt++ {
		db, err := gorm.Open(postgres.Open(cfg.DatabaseDSN), &gorm.Config{})
		if err == nil {
			log.Println("database connected")
			return db, nil
		}

		lastErr = err
		log.Printf("database connection attempt %d failed: %v", attempt, err)
		time.Sleep(2 * time.Second)
	}

	return nil, lastErr
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&productmodel.Product{},
		&ordermodel.Order{},
	)
}

func SeedProducts(db *gorm.DB) error {
	var count int64
	if err := db.Model(&productmodel.Product{}).Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	products := []productmodel.Product{
		{Name: "Laptop", Price: 1200, Stock: 5},
		{Name: "Keyboard", Price: 120, Stock: 15},
		{Name: "Mouse", Price: 80, Stock: 20},
	}

	return db.Create(&products).Error
}
