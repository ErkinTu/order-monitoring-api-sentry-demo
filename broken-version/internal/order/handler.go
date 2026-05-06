package order

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/example/order-monitoring-api-broken/internal/monitoring"
)

type Handler struct {
	service *Service
	repo    *Repository
}

func NewHandler(service *Service, repo *Repository) *Handler {
	return &Handler{service: service, repo: repo}
}

func (h *Handler) Create(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		monitoring.CaptureException(c, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		return
	}

	order, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		monitoring.CaptureException(c, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create order", "details": err.Error()})
		return
	}

	if req.SimulateSlow {
		monitoring.CaptureMessage(c, "slow order request was simulated")
	}

	c.JSON(http.StatusCreated, order)
}

func (h *Handler) CalculateDiscount(c *gin.Context) {
	var req DiscountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		monitoring.CaptureException(c, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		return
	}

	response := h.service.CalculateDiscount(req)
	c.JSON(http.StatusOK, response)
}

func (h *Handler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	order, err := h.repo.FindByID(c.Request.Context(), uint(id))
	if err != nil {
		monitoring.CaptureException(c, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load order", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, order)
}
