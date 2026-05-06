package router

import (
	"net/http"

	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	ordermodule "github.com/example/order-monitoring-api-broken/internal/order"
	productmodule "github.com/example/order-monitoring-api-broken/internal/product"
)

func New(db *gorm.DB) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(sentrygin.New(sentrygin.Options{
		Repanic:         true,
		WaitForDelivery: false,
	}))

	productRepo := productmodule.NewRepository(db)
	productHandler := productmodule.NewHandler(productRepo)

	orderRepo := ordermodule.NewRepository(db)
	orderService := ordermodule.NewService(orderRepo, productRepo)
	orderHandler := ordermodule.NewHandler(orderService, orderRepo)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.GET("/products", productHandler.List)
	r.POST("/orders", orderHandler.Create)
	r.GET("/orders/:id", orderHandler.GetByID)
	r.POST("/orders/discount", orderHandler.CalculateDiscount)

	return r
}
