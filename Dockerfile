# Build stage
FROM golang:1.22-alpine AS builder

# Install git and ca-certificates for HTTPS requests
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o commit-notifier ./cmd/commit-notifier

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN adduser -D -u 1000 notifier

# Copy binary from builder
COPY --from=builder /app/commit-notifier /usr/local/bin/commit-notifier

# Set ownership
RUN chmod +x /usr/local/bin/commit-notifier

# Switch to non-root user
USER notifier

# Run the notifier
ENTRYPOINT ["/usr/local/bin/commit-notifier"]