package handlers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"owngpt/models"
	"owngpt/services"
	"owngpt/utils"
)

type ModelHandler struct {
	dockerService *services.DockerService
	ollamaService *services.OllamaService
}

func NewModelHandler() *ModelHandler {
	return &ModelHandler{
		dockerService: services.NewDockerService(),
		ollamaService: services.NewOllamaService(),
	}
}

// CreateModel handles model creation requests
func (mh *ModelHandler) CreateModel(c *gin.Context) {
	var req models.CreateDockerfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Creating model: %s", req.Model)

	// Check if model is already running
	models.ModelMutex.RLock()
	if models.CurrentModel.IsRunning && strings.Contains(models.CurrentModel.Name, strings.ToLower(req.Model)) {
		models.ModelMutex.RUnlock()
		c.JSON(http.StatusOK, gin.H{
			"message":        "Model is already running and ready",
			"model":          req.Model,
			"container_name": models.CurrentModel.Name,
			"port":           models.CurrentModel.Port,
			"already_exists": true,
		})
		return
	}
	models.ModelMutex.RUnlock()

	// Check if model container already exists but stopped
	// Replace colons and other invalid characters in container names
	safeModelName := strings.ReplaceAll(strings.ToLower(req.Model), ":", "-")
	safeModelName = strings.ReplaceAll(safeModelName, "/", "-")
	containerName := fmt.Sprintf("ollama-%s-container", safeModelName)
	if mh.dockerService.ContainerExists(containerName) {
		log.Printf("Container %s already exists, starting it", containerName)
		if err := mh.dockerService.StartExistingContainer(containerName); err == nil {
			models.ModelMutex.Lock()
			models.CurrentModel = models.ModelContainer{
				Name:      containerName,
				Port:      "11434",
				IsRunning: true,
			}
			models.ModelMutex.Unlock()

			if err := mh.dockerService.WaitForModelReady(containerName, 30*time.Second); err == nil {
				c.JSON(http.StatusOK, gin.H{
					"message":        "Existing model container started successfully",
					"model":          req.Model,
					"container_name": containerName,
					"port":           "11434",
					"already_exists": true,
				})
				return
			}
		}
	}

	// Stop current model if running
	mh.stopCurrentModel()

	// Generate Dockerfile content
	dockerfileContent := utils.GenerateDockerfile(req.Model)

	// Create models directory if it doesn't exist
	modelsDir := "/app/models"
	if err := os.MkdirAll(modelsDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create models directory"})
		return
	}

	// Write Dockerfile
	dockerfilePath := filepath.Join(modelsDir, "Dockerfile")
	if err := os.WriteFile(dockerfilePath, []byte(dockerfileContent), 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write Dockerfile"})
		return
	}

	// Build Docker image
	imageName := fmt.Sprintf("ollama-%s", safeModelName)
	if err := mh.dockerService.BuildDockerImage(modelsDir, imageName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to build Docker image: %v", err)})
		return
	}

	// Run Docker container
	containerName = fmt.Sprintf("%s-container", imageName)
	port := "11434"
	if err := mh.dockerService.RunDockerContainer(imageName, containerName, port); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to run Docker container: %v", err)})
		return
	}

	// Update current model
	models.ModelMutex.Lock()
	models.CurrentModel = models.ModelContainer{
		Name:      containerName,
		Port:      port,
		IsRunning: true,
	}
	models.ModelMutex.Unlock()

	// Wait for the model to be ready
	if err := mh.dockerService.WaitForModelReady(containerName, 300*time.Second); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Model failed to start: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "Model created and container started successfully",
		"model":          req.Model,
		"container_name": containerName,
		"port":           port,
	})
}

// GetInstalledModels returns list of installed models
func (mh *ModelHandler) GetInstalledModels(c *gin.Context) {
	installedModels, err := mh.dockerService.GetInstalledModels()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list installed models"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"models": installedModels})
}

// GetAvailableModels returns list of available models
func (mh *ModelHandler) GetAvailableModels(c *gin.Context) {
	availableModels, err := mh.dockerService.GetAvailableModels()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get available models"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"available_models": availableModels})
}

// DeleteModel deletes a model and its container
func (mh *ModelHandler) DeleteModel(c *gin.Context) {
	modelName := c.Param("name")
	if modelName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Model name is required"})
		return
	}

	if err := mh.dockerService.DeleteModel(modelName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Update current model if it was the deleted one
	safeModelName := strings.ReplaceAll(strings.ToLower(modelName), ":", "-")
	safeModelName = strings.ReplaceAll(safeModelName, "/", "-")
	containerName := fmt.Sprintf("ollama-%s-container", safeModelName)
	models.ModelMutex.Lock()
	if models.CurrentModel.Name == containerName {
		models.CurrentModel = models.ModelContainer{}
	}
	models.ModelMutex.Unlock()

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Model %s deleted successfully", modelName)})
}

// GetSystemInfo returns system information including GPU availability
func (mh *ModelHandler) GetSystemInfo(c *gin.Context) {
	gpuAvailable := mh.dockerService.IsGPUAvailable()

	c.JSON(http.StatusOK, gin.H{
		"gpu_available": gpuAvailable,
		"memory_limit":  "4GB",
		"message": func() string {
			if gpuAvailable {
				return "GPU acceleration available - models will use GPU with 4GB memory limit"
			}
			return "CPU only - models will use CPU with 4GB memory limit"
		}(),
	})
}

// RefreshCurrentModel refreshes the current model state by detecting running containers
func (mh *ModelHandler) RefreshCurrentModel(c *gin.Context) {
	installedModels, err := mh.dockerService.GetInstalledModels()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to refresh model state"})
		return
	}

	// Find the first running model and set it as current
	models.ModelMutex.Lock()
	models.CurrentModel = models.ModelContainer{} // Reset current model
	for _, model := range installedModels {
		if model.IsRunning {
			models.CurrentModel = models.ModelContainer{
				Name:      model.ContainerName,
				Port:      "11434",
				IsRunning: true,
			}
			break
		}
	}
	currentModel := models.CurrentModel
	models.ModelMutex.Unlock()

	if currentModel.IsRunning {
		c.JSON(http.StatusOK, gin.H{
			"message":       "Current model refreshed successfully",
			"current_model": currentModel,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"message":       "No running models found",
			"current_model": nil,
		})
	}
}

// stopCurrentModel stops the currently running model
func (mh *ModelHandler) stopCurrentModel() {
	models.ModelMutex.Lock()
	defer models.ModelMutex.Unlock()

	if models.CurrentModel.IsRunning && models.CurrentModel.Name != "" {
		log.Printf("Stopping current model container: %s", models.CurrentModel.Name)
		// Note: We're not actually stopping it here, just marking as not current
		// The container will continue running but won't be the "current" model
		models.CurrentModel.IsRunning = false
	}
}
