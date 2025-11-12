# Build stage
FROM golang:1.23-alpine AS builder

# Go 환경 변수 설정
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies and verify
RUN go mod download && \
    go mod verify

# Copy source code
COPY main.go ./

# Build the application
# 에러 발생 시 더 명확한 메시지를 위해 단계별 실행
RUN go build \
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

