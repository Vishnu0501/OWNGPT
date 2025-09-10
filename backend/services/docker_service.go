package services

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"owngpt/models"
)

type DockerService struct{}

func NewDockerService() *DockerService {
	return &DockerService{}
}

// IsGPUAvailable checks if NVIDIA GPU is available for Docker
func (ds *DockerService) IsGPUAvailable() bool {
	// Check if nvidia-smi is available
	cmd := exec.Command("nvidia-smi")
	if err := cmd.Run(); err != nil {
		log.Printf("nvidia-smi not available: %v", err)
		return false
	}

	// Check if Docker supports GPU (nvidia-docker or Docker with GPU support)
	cmd = exec.Command("docker", "run", "--rm", "--gpus", "all", "hello-world")
	if err := cmd.Run(); err != nil {
		log.Printf("Docker GPU support not available: %v", err)
		return false
	}

	log.Println("GPU support detected and available")
	return true
}

// GetAvailableModels fetches available models from Docker Hub
func (ds *DockerService) GetAvailableModels() ([]models.AvailableModel, error) {
	// First, get popular hardcoded models for guaranteed availability
	popularModels := []models.AvailableModel{
		{Name: "mistral", Description: "Fast and efficient 7B model", Size: "4.1GB", Official: true},
		{Name: "llama2", Description: "Meta's powerful language model", Size: "3.8GB", Official: true},
		{Name: "llama2:13b", Description: "Larger Llama2 model with better performance", Size: "7.3GB", Official: true},
		{Name: "codellama", Description: "Specialized for code generation", Size: "3.8GB", Official: true},
		{Name: "codellama:13b", Description: "Larger CodeLlama for complex coding tasks", Size: "7.3GB", Official: true},
		{Name: "vicuna", Description: "Fine-tuned for conversations", Size: "3.8GB", Official: false},
		{Name: "orca-mini", Description: "Compact and fast model", Size: "1.9GB", Official: false},
		{Name: "neural-chat", Description: "Optimized for chat interactions", Size: "4.1GB", Official: false},
		{Name: "starcode", Description: "Code generation and completion", Size: "4.3GB", Official: false},
		{Name: "phind-codellama", Description: "Enhanced CodeLlama for development", Size: "3.8GB", Official: false},
	}

	// Try to get additional models from local Docker images
	localModels, err := ds.getLocalOllamaModels()
	if err == nil {
		// Merge local models with popular ones, avoiding duplicates
		modelMap := make(map[string]bool)
		for _, model := range popularModels {
			modelMap[model.Name] = true
		}

		for _, localModel := range localModels {
			if !modelMap[localModel.Name] {
				popularModels = append(popularModels, localModel)
			}
		}
	}

	return popularModels, nil
}

// getLocalOllamaModels gets models from local Docker images
func (ds *DockerService) getLocalOllamaModels() ([]models.AvailableModel, error) {
	cmd := exec.Command("docker", "images", "--format", "{{.Repository}}:{{.Tag}}\t{{.Size}}")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var localModels []models.AvailableModel
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		if strings.Contains(line, "ollama") && !strings.Contains(line, "ollama/ollama") {
			parts := strings.Split(line, "\t")
			if len(parts) >= 2 {
				imageName := parts[0]
				size := parts[1]

				// Extract model name from image name
				modelName := strings.TrimPrefix(imageName, "ollama-")
				modelName = strings.TrimSuffix(modelName, ":latest")

				if modelName != imageName { // Only if it's actually an ollama model
					localModels = append(localModels, models.AvailableModel{
						Name:        modelName,
						Description: "Locally available model",
						Size:        size,
						Official:    false,
					})
				}
			}
		}
	}

	return localModels, nil
}

// GetInstalledModels returns list of installed model containers
func (ds *DockerService) GetInstalledModels() ([]models.InstalledModel, error) {
	cmd := exec.Command("docker", "ps", "-a", "--format", "{{.Names}}\t{{.Status}}\t{{.Ports}}")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %v", err)
	}

	var installedModels []models.InstalledModel
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		if strings.Contains(line, "ollama-") && strings.Contains(line, "-container") {
			parts := strings.Split(line, "\t")
			if len(parts) >= 3 {
				containerName := parts[0]
				status := parts[1]
				ports := parts[2]

				// Extract model name
				modelName := strings.TrimSuffix(strings.TrimPrefix(containerName, "ollama-"), "-container")

				installedModels = append(installedModels, models.InstalledModel{
					Name:          modelName,
					ContainerName: containerName,
					Status:        status,
					Ports:         ports,
					IsRunning:     strings.Contains(status, "Up"),
				})
			}
		}
	}

	return installedModels, nil
}

// BuildDockerImage builds a Docker image for the specified model
func (ds *DockerService) BuildDockerImage(contextPath, imageName string) error {
	cmd := exec.Command("docker", "build", "-t", imageName, contextPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// RunDockerContainer runs a Docker container for the model
func (ds *DockerService) RunDockerContainer(imageName, containerName, port string) error {
	// Remove existing container if it exists
	exec.Command("docker", "rm", "-f", containerName).Run()

	// Base docker run arguments
	args := []string{
		"run", "-d", "--name", containerName,
		"--network", "owngpt_owngpt-network",
		"-p", fmt.Sprintf("%s:11434", port),
		"--restart", "unless-stopped",
		"--memory", "4g", // Limit memory to 4GB
	}

	// Add GPU support if available
	if ds.IsGPUAvailable() {
		args = append(args, "--gpus", "all")
		log.Printf("Starting container %s with GPU support and 4GB memory limit", containerName)
	} else {
		log.Printf("Starting container %s with CPU only and 4GB memory limit", containerName)
	}

	// Add the image name at the end
	args = append(args, imageName)

	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("Running command: docker %s\n", strings.Join(args, " "))
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Docker run failed: %v\n", err)
	}
	return err
}

// ContainerExists checks if a container exists
func (ds *DockerService) ContainerExists(containerName string) bool {
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

// StartExistingContainer starts an existing stopped container
func (ds *DockerService) StartExistingContainer(containerName string) error {
	cmd := exec.Command("docker", "start", containerName)
	return cmd.Run()
}

// DeleteModel removes a model container and image
func (ds *DockerService) DeleteModel(modelName string) error {
	safeModelName := strings.ReplaceAll(strings.ToLower(modelName), ":", "-")
	safeModelName = strings.ReplaceAll(safeModelName, "/", "-")
	containerName := fmt.Sprintf("ollama-%s-container", safeModelName)

	// Stop and remove the container
	cmd := exec.Command("docker", "rm", "-f", containerName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to remove container: %v", err)
	}

	// Remove the image
	imageName := fmt.Sprintf("ollama-%s", safeModelName)
	cmd = exec.Command("docker", "rmi", "-f", imageName)
	cmd.Run() // Don't fail if image removal fails

	return nil
}

// WaitForModelReady waits for the model container to be ready
func (ds *DockerService) WaitForModelReady(containerName string, timeout time.Duration) error {
	client := &http.Client{Timeout: 100 * time.Second}
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		// Use container name for internal Docker networking
		resp, err := client.Get(fmt.Sprintf("http://%s:11434/api/tags", containerName))
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			fmt.Println("Model is ready")
			return nil
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(2 * time.Second)
	}

	return fmt.Errorf("model failed to become ready within %v", timeout)
}
