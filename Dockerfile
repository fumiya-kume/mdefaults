# Build stage
FROM golang:latest AS builder

WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o mdefaults -ldflags="-s -w" .

# Final stage
FROM alpine:3.19

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/mdefaults /app/mdefaults

# Create a directory for configuration
RUN mkdir -p /root/.mdefaults

# Add a warning message
RUN echo "WARNING: mdefaults is designed for macOS and may not function correctly in this Docker container." > /app/WARNING.txt

# Set the entrypoint
ENTRYPOINT ["/app/mdefaults"]
