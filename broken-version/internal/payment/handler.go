package payment

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/example/order-monitoring-api-broken/internal/monitoring"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Capture(c *gin.Context) {
	orderID, ok := parseOrderID(c)
	if !ok {
		return
	}

	payment, err := h.service.Capture(c.Request.Context(), orderID)
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		case errors.Is(err, ErrOrderAlreadyPaid):
			c.JSON(http.StatusConflict, gin.H{"error": "order already paid"})
		default:
			monitoring.CaptureException(c, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to capture payment", "details": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, payment)
}

func (h *Handler) ListByOrderID(c *gin.Context) {
	orderID, ok := parseOrderID(c)
	if !ok {
		return
	}

	payments, err := h.service.ListByOrderID(c.Request.Context(), orderID)
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		default:
			monitoring.CaptureException(c, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load payments", "details": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, payments)
}

func parseOrderID(c *gin.Context) (uint, bool) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return 0, false
	}

	return uint(id), true
}
