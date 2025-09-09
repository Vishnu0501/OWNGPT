FROM node:18-alpine

WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm ci

# Copy source code
COPY . .

# Build the app
RUN npm run build

# Install a simple static file server
RUN npm install -g serve

# Expose port
EXPOSE 9090

# Serve the built app
CMD ["serve", "-s", "dist", "-l", "9090"]
