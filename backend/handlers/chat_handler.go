package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"owngpt/models"
	"owngpt/services"
)

type ChatHandler struct {
	ollamaService *services.OllamaService
}

func NewChatHandler() *ChatHandler {
	return &ChatHandler{
		ollamaService: services.NewOllamaService(),
	}
}

// SendMessageStream handles streaming chat message requests
func (ch *ChatHandler) SendMessageStream(c *gin.Context) {
	var req models.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	models.ModelMutex.RLock()
	if !models.CurrentModel.IsRunning {
		models.ModelMutex.RUnlock()
		c.JSON(http.StatusBadRequest, gin.H{"error": "No model is currently running. Please create a model first."})
		return
	}
	containerName := models.CurrentModel.Name
	models.ModelMutex.RUnlock()

	log.Printf("Streaming message to model: %s", req.Message)

	// Set headers for Server-Sent Events
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// Get streaming response
	responseChan, errorChan := ch.ollamaService.SendMessageStream(req.Message, containerName)

	// Stream responses to client
	for {
		select {
		case response, ok := <-responseChan:
			if !ok {
				return
			}
			if response != "" {
				c.SSEvent("data", response)
				c.Writer.Flush()
			}
		case err := <-errorChan:
			if err != nil {
				c.SSEvent("error", fmt.Sprintf("Error: %v", err))
				c.Writer.Flush()
			}
			return
		}
	}
}

// SendMessage handles chat message requests
func (ch *ChatHandler) SendMessage(c *gin.Context) {
	var req models.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	models.ModelMutex.RLock()
	if !models.CurrentModel.IsRunning {
		models.ModelMutex.RUnlock()
		c.JSON(http.StatusBadRequest, gin.H{"error": "No model is currently running. Please create a model first."})
		return
	}
	containerName := models.CurrentModel.Name
	models.ModelMutex.RUnlock()

	log.Printf("Sending message to model: %s", req.Message)

	// Send message to Ollama
	response, err := ch.ollamaService.SendMessage(req.Message, containerName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ChatResponse{
			Error: fmt.Sprintf("Failed to get response from model: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, models.ChatResponse{
		Response: response,
	})
}
