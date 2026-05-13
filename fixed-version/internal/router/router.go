package router

import (
	"net/http"

	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/example/order-monitoring-api-fixed/internal/logging"
	ordermodule "github.com/example/order-monitoring-api-fixed/internal/order"
	paymentmodule "github.com/example/order-monitoring-api-fixed/internal/payment"
	productmodule "github.com/example/order-monitoring-api-fixed/internal/product"
)

func New(db *gorm.DB) *gin.Engine {
	r := gin.New()
	r.Use(logging.GinMiddleware())
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

	paymentRepo := paymentmodule.NewRepository(db)
	paymentService := paymentmodule.NewService(db, paymentRepo, orderRepo)
	paymentHandler := paymentmodule.NewHandler(paymentService)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.GET("/products", productHandler.List)
	r.POST("/orders", orderHandler.Create)
	r.GET("/orders/:id", orderHandler.GetByID)
	r.POST("/orders/discount", orderHandler.CalculateDiscount)
	r.POST("/orders/:id/payments", paymentHandler.Capture)
	r.GET("/orders/:id/payments", paymentHandler.ListByOrderID)

	return r
}
