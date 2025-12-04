package handlers

import (
	"net/http"
	"github.com/ollatomiwa/hotelsystem/payment-service/internal/models"
	"github.com/ollatomiwa/hotelsystem/payment-service/pkg/database"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	db *database.Database
}

func NewHealthHandler(db *database.Database) *HealthHandler {
	return &HealthHandler{db: db}
}

// HealthCheck godoc
// @Summary Health check
// @Description Check service health status
// @Tags health
// @Produce json
// @Success 200 {object} models.APIResponse
// @Router /api/v1/health [get]
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	status := "healthy"
	dbStatus := "ok"

	// Check database connection
	if err := h.db.HealthCheck(); err != nil {
		dbStatus = "error: " + err.Error()
		status = "unhealthy"
	}

	data := gin.H{
		"status":   status,
		"database": dbStatus,
	}

	httpStatus := http.StatusOK
	if status == "unhealthy" {
		httpStatus = http.StatusServiceUnavailable
	}

	c.JSON(httpStatus, models.APIResponse{
		Status:  status,
		Message: "Health check completed",
		Data:    data,
	})
}