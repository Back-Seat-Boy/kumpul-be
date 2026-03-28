# Build stage
FROM golang:1.26.1-alpine AS builder

WORKDIR /app

# Install git and ca-certificates for go mod download
RUN apk add --no-cache git ca-certificates

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/kumpul main.go

# Final stage
FROM alpine:latest

WORKDIR /app

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Copy binary from builder
COPY --from=builder /app/bin/kumpul /app/kumpul

# Copy migrations
COPY --from=builder /app/db/migrations /app/db/migrations

# Copy config example (will be overridden by volume or env)
COPY --from=builder /app/config.yml.example /app/config.yml.example

# Expose port
EXPOSE 8080

# Run the binary
CMD ["./kumpul", "server"]
