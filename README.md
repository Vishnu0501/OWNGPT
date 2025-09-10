# OwnGPT - Personal AI Assistant

A containerized chatbot application that allows you to run various AI models locally using Ollama, with a Go backend and React frontend.

## 🚀 Features

- **Dynamic Model Loading**: Create and run any Ollama model on-demand
- **Modern UI**: Beautiful React-based chat interface with model management
- **Interactive Chat**: Chat with any running model through an intuitive interface
- **Custom Model Support**: Pull and create any model from Ollama library with text input
- **Model Management**: Easy model selection, installation, and switching
- **Containerized**: Everything runs in Docker containers for easy deployment
- **RESTful API**: Clean Go backend with Gin framework
- **Real-time Chat**: Seamless conversation flow with AI models

## 🏗️ Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   React Frontend │    │   Go Backend    │    │  Ollama Model   │
│   (Port 9090)    │◄──►│   (Port 8080)   │◄──►│  (Port 11434)   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 📋 Prerequisites

- Docker and Docker Compose installed
- At least 4GB of RAM (for running AI models)
- Internet connection (for downloading models)

## 🚀 Quick Start

1. **Clone and navigate to the project**:
   ```bash
   git clone <your-repo-url>
   cd OWNGPT
   ```

2. **Start the application**:
   ```bash
   docker-compose up --build
   ```

3. **Access the application**:
   - Frontend: http://localhost:9090
   - Backend API: http://localhost:8080

## 📖 Usage

### Model Management
1. **Browse Available Models**: Switch to the "Available Models" tab to see popular models
2. **Create Custom Models**: 
   - Type any model name in the custom model input field (e.g., `llama3`, `gemma`, `phi3`, `qwen`)
   - Use the suggestion tags for quick selection of popular models
   - Press Enter or click "Pull & Create" to install the model
   - Wait for the model to download and start (this may take a few minutes)
3. **View Installed Models**: Switch to "Installed Models" tab to see all your models and their status
4. **Quick Chat**: Click the "Quick Chat" button in the header when models are running

### Chat Interface
1. **Start Chat**: Click "💬 Chat Now" on any running model from the installed models list
2. **Switch Models**: Use the "🔧 Switch Model" button to go back to model selection
3. **Clear Chat**: Use "🗑️ Clear Chat" to start a fresh conversation
4. **Model Status**: See the current model name and running status in the header

### Supported Models
You can install any model available in the Ollama library, including:
- **llama3**: Latest Llama model from Meta
- **gemma**: Google's Gemma models
- **phi3**: Microsoft's Phi-3 models  
- **qwen**: Alibaba's Qwen models
- **codellama**: Code-specialized Llama model
- **mistral**: Mistral AI models
- And many more! Just type the model name in the custom input field.

## 🔧 API Endpoints

### POST /create-dockerfile
Creates and runs a new Ollama model container.

**Request:**
```json
{
  "model": "mistral"
}
```

**Response:**
```json
{
  "message": "Dockerfile created and container started successfully",
  "model": "mistral",
  "container_name": "ollama-mistral-container",
  "port": "11434"
}
```

### POST /chat
Sends a message to the running model.

**Request:**
```json
{
  "message": "Hello, how are you?"
}
```

**Response:**
```json
{
  "response": "Hello! I'm doing well, thank you for asking. How can I help you today?"
}
```

### GET /health
Returns the health status of the backend and current model.

**Response:**
```json
{
  "status": "healthy",
  "model_running": true,
  "model_name": "ollama-mistral-container"
}
```

## 🐳 Docker Services

- **Backend**: Go application with Gin framework
- **Frontend**: React application served by Node.js static server
- **Model**: Dynamically created Ollama containers

## 🛠️ Development

### Backend Development
```bash
cd backend
go mod tidy
go run main.go
```

### Frontend Development
```bash
cd frontend
npm install
npm run dev
```

## 📁 Project Structure

```
OWNGPT/
├── backend/
│   ├── main.go              # Main Go application
│   ├── go.mod               # Go dependencies
│   ├── go.sum               # Go checksums
│   └── Dockerfile           # Backend container
├── frontend/
│   ├── src/
│   │   ├── App.jsx          # Main React component
│   │   ├── main.jsx         # React entry point
│   │   └── index.css        # Styles
│   ├── package.json         # Node dependencies
│   ├── vite.config.js       # Vite configuration
│   └── Dockerfile           # Frontend container
├── docker-compose.yml       # Orchestration
└── README.md               # This file
```

## 🔧 Configuration

### Environment Variables
- `BACKEND_PORT`: Backend server port (default: 8080)
- `FRONTEND_PORT`: Frontend server port (default: 9090)
- `GIN_MODE`: Gin framework mode (default: release)

### Supported Models
Any model available in Ollama Hub:
- mistral
- llama2
- codellama
- vicuna
- orca-mini
- And many more...

## 🚨 Troubleshooting

### Model Creation Issues
- Ensure Docker has enough memory allocated
- Check internet connection for model downloads
- Verify Docker daemon is running

### Connection Issues
- Make sure all containers are running: `docker-compose ps`
- Check logs: `docker-compose logs <service-name>`
- Verify ports are not in use by other applications

### Performance Issues
- Increase Docker memory limits
- Use smaller models for better performance
- Ensure sufficient disk space for model downloads

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

## 📝 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🙏 Acknowledgments

- [Ollama](https://ollama.ai/) for providing the model runtime
- [Gin](https://gin-gonic.com/) for the Go web framework
- [React](https://reactjs.org/) for the frontend framework
- [Vite](https://vitejs.dev/) for the build tool
