package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"owngpt/models"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// CheckHealth returns the health status of the application
func (hh *HealthHandler) CheckHealth(c *gin.Context) {
	models.ModelMutex.RLock()
	defer models.ModelMutex.RUnlock()

	c.JSON(http.StatusOK, gin.H{
		"status":        "healthy",
		"model_running": models.CurrentModel.IsRunning,
		"model_name":    models.CurrentModel.Name,
	})
}
