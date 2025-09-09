# OwnGPT - Personal AI Assistant

A containerized chatbot application that allows you to run various AI models locally using Ollama, with a Go backend and React frontend.

## 🚀 Features

- **Dynamic Model Loading**: Create and run any Ollama model on-demand
- **Modern UI**: Beautiful React-based chat interface
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

1. **Create a Model**:
   - Enter a model name (e.g., "mistral", "llama2", "codellama")
   - Click "Create Model"
   - Wait for the model to download and start (this may take a few minutes)

2. **Start Chatting**:
   - Once the model is ready, type your message
   - Press Enter or click "Send"
   - Enjoy your conversation with the AI!

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
