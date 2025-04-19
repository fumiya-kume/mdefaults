FROM golang:1.23-alpine AS builder

WORKDIR /app

# Download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code and build
COPY . .
RUN CGO_ENABLED=0 go build -o mdefaults ./cmd/mdefaults

FROM alpine:latest
WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/mdefaults /app/mdefaults

# Set the default entrypoint
ENTRYPOINT ["/app/mdefaults"]