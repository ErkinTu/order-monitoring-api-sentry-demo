package payment

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/example/order-monitoring-api-fixed/internal/monitoring"
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
		handleServiceError(c, err)
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
		handleServiceError(c, err)
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

func handleServiceError(c *gin.Context, err error) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		if appErr.Status >= 500 {
			monitoring.CaptureException(c, err)
		}
		c.JSON(appErr.Status, gin.H{"error": appErr.Message})
		return
	}

	monitoring.CaptureException(c, err)
	c.JSON(http.StatusInternalServerError, gin.H{"error": "unexpected server error"})
}
