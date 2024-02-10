package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AddAccount godoc
// @Summary      API Health
// @Description  API Health response
// @Tags         health
// @Accept       json
// @Produce      json
// @Param        user  body      model.HealthResponse  true  "Health"
// @Success      200      {object}  model.HealthResponse
// @Router       /health [post]
func Ready(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{
		Status: "ready",
	})
}
