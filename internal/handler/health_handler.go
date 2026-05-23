package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthResponse describes the /health payload.
type HealthResponse struct {
	Status string `json:"status" example:"ok"`
}

// Health godoc
// @Summary      Health check
// @Description  Liveness probe. Returns 200 OK when the service is running.
// @Tags         health
// @Produce      json
// @Success      200  {object}  HealthResponse
// @Router       /health [get]
func Health(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{Status: "ok"})
}
