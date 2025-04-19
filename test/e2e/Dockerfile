FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 go build -o mdefaults ./cmd/mdefaults

# Create a minimal image for running the application
FROM alpine:latest

WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/mdefaults /app/mdefaults

# Set the entrypoint to the mdefaults binary
ENTRYPOINT ["/app/mdefaults"]
