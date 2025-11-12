# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY main.go ./

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
    -ldflags="-s -w" \
    -trimpath \
    -o seal-automate main.go

# Runtime stage
FROM alpine:latest

# Install CA certificates for TLS
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/seal-automate .

# Create directory for TLS certificates
RUN mkdir -p /tls

# Expose port
EXPOSE 8443

# Run the application
CMD ["./seal-automate"]

