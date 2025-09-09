FROM golang:1.21-alpine AS builder

# Install dependencies for building
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM alpine:latest

# Install docker client and other necessary tools
RUN apk --no-cache add ca-certificates docker-cli bash curl

# Create app directory
WORKDIR /app

# Create models directory
RUN mkdir -p /app/models

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Expose port
EXPOSE 8080

# Run the binary
CMD ["./main"]
