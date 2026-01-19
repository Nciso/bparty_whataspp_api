# Build stage
FROM golang:1.23-alpine AS builder

# Install SQLite and build dependencies
RUN apk add --no-cache sqlite sqlite-dev gcc musl-dev

# Set working directory
WORKDIR /app

# Enable automatic toolchain download for Go 1.24+
ENV GOTOOLCHAIN=auto

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o whatsapp-api main.go

# Runtime stage
FROM alpine:latest

# Install SQLite runtime (needed for go-sqlite3)
RUN apk add --no-cache sqlite ca-certificates

WORKDIR /app

# Copy binary from builder to both /app and /usr/local/bin
COPY --from=builder /app/whatsapp-api /app/whatsapp-api
COPY --from=builder /app/whatsapp-api /usr/local/bin/whatsapp-api

# Ensure binaries are executable and verify they exist
RUN chmod +x /app/whatsapp-api /usr/local/bin/whatsapp-api && \
    ls -la /app/ && \
    ls -la /usr/local/bin/whatsapp-api && \
    test -f /app/whatsapp-api && echo "Binary in /app verified" && \
    test -f /usr/local/bin/whatsapp-api && echo "Binary in PATH verified"

# Expose port
EXPOSE 8080

# Use ENTRYPOINT so Railway can find it
ENTRYPOINT ["whatsapp-api"]
