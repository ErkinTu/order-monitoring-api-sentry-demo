package product

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/example/order-monitoring-api-fixed/internal/monitoring"
)

type Handler struct {
	repo *Repository
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) List(c *gin.Context) {
	products, err := h.repo.List(c.Request.Context())
	if err != nil {
		monitoring.CaptureException(c, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load products"})
		return
	}

	c.JSON(http.StatusOK, products)
}
