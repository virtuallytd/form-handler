# Use the official Golang image to create a build artifact.
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY app/go.mod app/go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code
COPY app/ .

# Build the Go app with CGO enabled
RUN CGO_ENABLED=1 GOOS=linux go build -o main .

# Start a new stage from scratch
FROM alpine:latest

# Create a non-root user and group
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache sqlite curl

# Create necessary directories
RUN mkdir -p /app/uploads /app/logs /app/backend/static /app/config

# Copy the pre-built binary and other necessary files
COPY --from=builder /app/main .
COPY config/config.json /app/config/config.json
COPY app/backend/ /app/backend/
COPY app/backend/tailwind.min.css /app/backend/static/tailwind.min.css
COPY .env /app/

# Change ownership of the directories to the non-root user
RUN chown -R appuser:appgroup /app/

# Expose port 8080 to the outside world
EXPOSE 8080

# Health check to ensure the container is healthy
HEALTHCHECK --interval=30s --timeout=10s --start-period=30s CMD curl -f http://localhost:8080/health || exit 1

# Switch to the non-root user
USER appuser

# Command to run the executable
CMD ["./main"]
