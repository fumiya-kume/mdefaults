FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o mdefaults ./cmd/mdefaults

# Create a minimal image for testing
FROM alpine:latest

WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/mdefaults /app/mdefaults

# Copy the e2e test scripts
COPY test/e2e/*.sh /app/

# Make the scripts executable
RUN chmod +x /app/*.sh

# Set the entrypoint to the test script
ENTRYPOINT ["/app/run_tests.sh"]