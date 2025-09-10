package routes

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"owngpt/handlers"
)

// SetupRoutes configures all the routes for the application
func SetupRoutes() *gin.Engine {
	r := gin.Default()

	// Configure CORS
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:9090", "http://frontend:9090"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	r.Use(cors.New(config))

	// Initialize handlers
	modelHandler := handlers.NewModelHandler()
	chatHandler := handlers.NewChatHandler()
	healthHandler := handlers.NewHealthHandler()

	// Health routes
	r.GET("/health", healthHandler.CheckHealth)

	// Model management routes
	r.POST("/create-dockerfile", modelHandler.CreateModel)
	r.GET("/models", modelHandler.GetInstalledModels)
	r.GET("/available-models", modelHandler.GetAvailableModels)
	r.DELETE("/models/:name", modelHandler.DeleteModel)
	r.POST("/refresh-model", modelHandler.RefreshCurrentModel)
	r.GET("/system-info", modelHandler.GetSystemInfo)

	// Chat routes
	r.POST("/chat", chatHandler.SendMessage)
	r.POST("/chat/stream", chatHandler.SendMessageStream)

	return r
}
