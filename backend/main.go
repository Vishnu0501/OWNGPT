package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type CreateDockerfileRequest struct {
	Model string `json:"model" binding:"required"`
}

type ChatRequest struct {
	Message string `json:"message" binding:"required"`
}

type ChatResponse struct {
	Response string `json:"response"`
	Error    string `json:"error,omitempty"`
}

type OllamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

type ModelContainer struct {
	Name      string
	Port      string
	IsRunning bool
}

var (
	currentModel ModelContainer
	modelMutex   sync.RWMutex
)

func main() {
	// Create gin router
	r := gin.Default()

	// Configure CORS
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:9090", "http://frontend:9090"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	r.Use(cors.New(config))

	// Routes
	r.POST("/create-dockerfile", createDockerfileHandler)
	r.POST("/chat", chatHandler)
	r.GET("/health", healthHandler)

	// Start server
	log.Println("Starting server on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func createDockerfileHandler(c *gin.Context) {
	var req CreateDockerfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Creating Dockerfile for model: %s", req.Model)

	// Check if model is already running
	modelMutex.RLock()
	if currentModel.IsRunning && strings.Contains(currentModel.Name, strings.ToLower(req.Model)) {
		modelMutex.RUnlock()
		c.JSON(http.StatusOK, gin.H{
			"message":        "Model is already running and ready",
			"model":          req.Model,
			"container_name": currentModel.Name,
			"port":           currentModel.Port,
			"already_exists": true,
		})
		return
	}
	modelMutex.RUnlock()

	// Check if model container already exists but stopped
	containerName := fmt.Sprintf("ollama-%s-container", strings.ToLower(req.Model))
	if containerExists(containerName) {
		log.Printf("Container %s already exists, starting it", containerName)
		if err := startExistingContainer(containerName); err == nil {
			modelMutex.Lock()
			currentModel = ModelContainer{
				Name:      containerName,
				Port:      "11434",
				IsRunning: true,
			}
			modelMutex.Unlock()

			if err := waitForModelReady(containerName, 30*time.Second); err == nil {
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
	stopCurrentModel()

	// Generate Dockerfile content
	dockerfileContent := generateDockerfile(req.Model)

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
	imageName := fmt.Sprintf("ollama-%s", strings.ToLower(req.Model))
	if err := buildDockerImage(modelsDir, imageName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to build Docker image: %v", err)})
		return
	}

	// Run Docker container
	containerName = fmt.Sprintf("%s-container", imageName)
	port := "11434"
	if err := runDockerContainer(imageName, containerName, port); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to run Docker container: %v", err)})
		return
	}

	// Update current model
	modelMutex.Lock()
	currentModel = ModelContainer{
		Name:      containerName,
		Port:      port,
		IsRunning: true,
	}
	modelMutex.Unlock()

	// Wait for the model to be ready
	if err := waitForModelReady(containerName, 300*time.Second); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Model failed to start: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "Dockerfile created and container started successfully",
		"model":          req.Model,
		"container_name": containerName,
		"port":           port,
	})
}

func chatHandler(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	modelMutex.RLock()
	if !currentModel.IsRunning {
		modelMutex.RUnlock()
		c.JSON(http.StatusBadRequest, gin.H{"error": "No model is currently running. Please create a model first."})
		return
	}
	containerName := currentModel.Name
	modelMutex.RUnlock()

	log.Printf("Sending message to model: %s", req.Message)

	// Send message to Ollama
	response, err := sendToOllama(req.Message, containerName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ChatResponse{
			Error: fmt.Sprintf("Failed to get response from model: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, ChatResponse{
		Response: response,
	})
}

func healthHandler(c *gin.Context) {
	modelMutex.RLock()
	defer modelMutex.RUnlock()

	c.JSON(http.StatusOK, gin.H{
		"status":        "healthy",
		"model_running": currentModel.IsRunning,
		"model_name":    currentModel.Name,
	})
}

func generateDockerfile(model string) string {
	return fmt.Sprintf(`FROM ollama/ollama:latest

# Install curl for health checks
RUN apt-get update && apt-get install -y curl && rm -rf /var/lib/apt/lists/*

# Expose Ollama port
EXPOSE 11434

# Create startup script
RUN echo '#!/bin/bash\n\
set -e\n\
echo "Starting Ollama server..."\n\
ollama serve &\n\
OLLAMA_PID=$!\n\
\n\
echo "Waiting for Ollama to be ready..."\n\
sleep 10\n\
while ! curl -s http://localhost:11434/api/tags >/dev/null 2>&1; do\n\
    sleep 2\n\
    echo "Still waiting for Ollama..."\n\
done\n\
\n\
echo "Ollama is ready, pulling model: %s"\n\
ollama pull %s\n\
\n\
echo "Model %s is ready!"\n\
wait $OLLAMA_PID' > /start.sh && chmod +x /start.sh

# Use bash to run the script
CMD ["/bin/bash", "/start.sh"]
`, strings.ToLower(model), strings.ToLower(model), strings.ToLower(model))
}

func buildDockerImage(contextPath, imageName string) error {
	cmd := exec.Command("docker", "build", "-t", imageName, contextPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runDockerContainer(imageName, containerName, port string) error {
	// Remove existing container if it exists
	exec.Command("docker", "rm", "-f", containerName).Run()

	// Run new container connected to the same network as the backend
	cmd := exec.Command("docker", "run", "-d", "--name", containerName,
		"--network", "owngpt_owngpt-network",
		"-p", fmt.Sprintf("%s:11434", port),
		imageName)

	// Capture both stdout and stderr for debugging
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("Running command: %v", cmd.Args)
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Docker run failed: %v", err)
	}
	return err
}

func stopCurrentModel() {
	modelMutex.Lock()
	defer modelMutex.Unlock()

	if currentModel.IsRunning && currentModel.Name != "" {
		log.Printf("Stopping current model container: %s", currentModel.Name)
		cmd := exec.Command("docker", "rm", "-f", currentModel.Name)
		cmd.Run()
		currentModel.IsRunning = false
	}
}

func waitForModelReady(containerName string, timeout time.Duration) error {
	client := &http.Client{Timeout: 50 * time.Second}
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		// Use container name for internal Docker networking
		resp, err := client.Get(fmt.Sprintf("http://%s:11434/api/tags", containerName))
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			log.Println("Model is ready")
			return nil
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(2 * time.Second)
	}

	return fmt.Errorf("model failed to become ready within %v", timeout)
}

func containerExists(containerName string) bool {
	cmd := exec.Command("docker", "ps", "-a", "--format", "{{.Names}}")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	containers := strings.Split(string(output), "\n")
	for _, container := range containers {
		if strings.TrimSpace(container) == containerName {
			return true
		}
	}
	return false
}

func startExistingContainer(containerName string) error {
	cmd := exec.Command("docker", "start", containerName)
	return cmd.Run()
}

func sendToOllama(message, containerName string) (string, error) {
	client := &http.Client{Timeout: 300 * time.Second}

	// Extract model name from container name
	modelName := strings.TrimSuffix(strings.TrimPrefix(containerName, "ollama-"), "-container")

	payload := map[string]interface{}{
		"model":  modelName,
		"prompt": message,
		"stream": false,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	// Use container name for internal Docker networking
	url := fmt.Sprintf("http://%s:11434/api/generate", containerName)
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var ollamaResp OllamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return "", err
	}

	return ollamaResp.Response, nil
}
