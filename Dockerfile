# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev sqlite-dev

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application with SQLite support
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o carnac .

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates sqlite-libs sqlite

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/carnac .

# Create data directory for SQLite
RUN mkdir -p /data

# Expose port
EXPOSE 8080

# Run the application
CMD ["./carnac", "-server", "-port", "8080"]
