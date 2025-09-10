package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"owngpt/models"
)

type OllamaService struct{}

func NewOllamaService() *OllamaService {
	return &OllamaService{}
}

// SendMessage sends a message to the Ollama model and returns the response
func (os *OllamaService) SendMessage(message, containerName string) (string, error) {
	// Optimized HTTP client with connection pooling and aggressive timeout
	client := &http.Client{
		Timeout: 15 * time.Second, // Aggressive timeout for sub-6s responses
		Transport: &http.Transport{
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     30 * time.Second,
		},
	}

	// Extract model name from container name
	modelName := strings.TrimSuffix(strings.TrimPrefix(containerName, "ollama-"), "-container")

	// Optimized payload with performance parameters
	payload := map[string]interface{}{
		"model":  modelName,
		"prompt": message,
		"stream": false,
		"options": map[string]interface{}{
			"num_predict":    250,   // Reduced for sub-6s responses
			"temperature":    0.2,   // Much lower for faster, focused responses
			"top_p":          0.7,   // More focused sampling
			"top_k":          15,    // Limit vocabulary for speed
			"num_ctx":        512,   // Much smaller context for speed
			"num_batch":      128,   // Smaller batch for faster processing
			"num_gpu":        1,     // Use GPU if available
			"low_vram":       false, // Don't limit VRAM usage for speed
			"f16_kv":         true,  // Use FP16 for key-value cache (faster)
			"use_mlock":      true,  // Keep model in memory
			"use_mmap":       true,  // Memory-mapped model loading
			"repeat_penalty": 1.05,  // Minimal penalty for speed
			"tfs_z":          0.95,  // Tail free sampling for speed
		},
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

	var ollamaResp models.OllamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return "", err
	}

	return ollamaResp.Response, nil
}

// SendMessageStream sends a message and returns streaming response for faster UI updates
func (os *OllamaService) SendMessageStream(message, containerName string) (chan string, chan error) {
	responseChan := make(chan string, 10)
	errorChan := make(chan error, 1)

	go func() {
		defer close(responseChan)
		defer close(errorChan)

		// Optimized HTTP client for streaming
		client := &http.Client{
			Timeout: 15 * time.Second, // Aggressive timeout for sub-6s responses
			Transport: &http.Transport{
				MaxIdleConns:        10,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     30 * time.Second,
			},
		}

		// Extract model name from container name
		modelName := strings.TrimSuffix(strings.TrimPrefix(containerName, "ollama-"), "-container")

		// Streaming payload with optimized parameters
		payload := map[string]interface{}{
			"model":  modelName,
			"prompt": message,
			"stream": true, // Enable streaming
			"options": map[string]interface{}{
				"num_predict":    250,   // Reduced for sub-6s responses
				"temperature":    0.2,   // Much lower for faster responses
				"top_p":          0.7,   // More focused sampling
				"top_k":          15,    // Limit vocabulary for speed
				"num_ctx":        512,   // Much smaller context for speed
				"num_batch":      128,   // Smaller batch for faster processing
				"num_gpu":        1,     // Use GPU if available
				"low_vram":       false, // Don't limit VRAM usage for speed
				"f16_kv":         true,  // Use FP16 for key-value cache (faster)
				"use_mlock":      true,  // Keep model in memory
				"use_mmap":       true,  // Memory-mapped model loading
				"repeat_penalty": 1.05,  // Minimal penalty for speed
				"tfs_z":          0.95,  // Tail free sampling for speed
			},
		}

		jsonData, err := json.Marshal(payload)
		if err != nil {
			errorChan <- err
			return
		}

		url := fmt.Sprintf("http://%s:11434/api/generate", containerName)
		resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			errorChan <- err
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			errorChan <- fmt.Errorf("ollama API returned status %d: %s", resp.StatusCode, string(body))
			return
		}

		// Read streaming response line by line
		decoder := json.NewDecoder(resp.Body)
		var fullResponse strings.Builder

		for decoder.More() {
			var streamResp models.OllamaResponse
			if err := decoder.Decode(&streamResp); err != nil {
				errorChan <- err
				return
			}

			if streamResp.Response != "" {
				fullResponse.WriteString(streamResp.Response)
				responseChan <- streamResp.Response
			}

			if streamResp.Done {
				break
			}
		}

		// Send final complete response
		responseChan <- fullResponse.String()
	}()

	return responseChan, errorChan
}
