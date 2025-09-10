package utils

import (
	"fmt"
	"strings"
)

// GenerateDockerfile generates a Dockerfile content for the specified model

func GenerateDockerfile(model string) string {
	return fmt.Sprintf(`FROM ollama/ollama:latest

# Install curl for health checks
RUN apt-get update && apt-get install -y curl && rm -rf /var/lib/apt/lists/*

# Set aggressive performance environment variables for sub-6s responses
ENV OLLAMA_NUM_PARALLEL=2
ENV OLLAMA_MAX_LOADED_MODELS=1
ENV OLLAMA_FLASH_ATTENTION=1
ENV OLLAMA_LLM_LIBRARY=cpu
ENV OLLAMA_KEEP_ALIVE=10m
ENV OLLAMA_HOST=0.0.0.0:11434
ENV OLLAMA_MAX_QUEUE=1
ENV OLLAMA_RUNNERS_DIR=/tmp

# Expose Ollama port
EXPOSE 11434

# Create optimized startup script
RUN echo '#!/bin/bash\n\
set -e\n\
echo "Starting optimized Ollama server..."\n\
\n\
# Set aggressive performance options for sub-6s responses\n\
export OLLAMA_NUM_PARALLEL=2\n\
export OLLAMA_MAX_LOADED_MODELS=1\n\
export OLLAMA_FLASH_ATTENTION=1\n\
export OLLAMA_KEEP_ALIVE=10m\n\
export OLLAMA_HOST=0.0.0.0:11434\n\
export OLLAMA_MAX_QUEUE=1\n\
export OLLAMA_RUNNERS_DIR=/tmp\n\
\n\
# Start Ollama with optimizations\n\
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
echo "Preloading model for faster responses..."\n\
curl -X POST http://localhost:11434/api/generate -d '"'"'{"model": "%s", "prompt": "Hello", "stream": false, "keep_alive": "5m"}'"'"' || true\n\
\n\
echo "Model %s is ready and optimized!"\n\
wait $OLLAMA_PID' > /usr/local/bin/start-with-model.sh && chmod +x /usr/local/bin/start-with-model.sh

# Override the entrypoint to use our script
ENTRYPOINT ["/usr/local/bin/start-with-model.sh"]
`, strings.ToLower(model), strings.ToLower(model), strings.ToLower(model), strings.ToLower(model))
}
