package main

import (
	"log"

	"owngpt/models"
	"owngpt/routes"
	"owngpt/services"
)

func main() {
	// Initialize model detection on startup
	initializeCurrentModel()

	// Setup routes
	r := routes.SetupRoutes()

	// Start server
	log.Println("Starting OwnGPT server on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

// initializeCurrentModel detects any running model containers on startup
func initializeCurrentModel() {
	dockerService := services.NewDockerService()
	installedModels, err := dockerService.GetInstalledModels()
	if err != nil {
		log.Printf("Failed to check for existing models: %v", err)
		return
	}

	// Find the first running model and set it as current
	for _, model := range installedModels {
		if model.IsRunning {
			models.ModelMutex.Lock()
			models.CurrentModel = models.ModelContainer{
				Name:      model.ContainerName,
				Port:      "11434", // Default Ollama port
				IsRunning: true,
			}
			models.ModelMutex.Unlock()
			log.Printf("Detected running model: %s (container: %s)", model.Name, model.ContainerName)
			return
		}
	}

	log.Println("No running models detected on startup")
}
