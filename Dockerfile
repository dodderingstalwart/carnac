# Build stage
FROM golang:1.24-alpine AS builder

# Install necessary build dependencies
RUN apk add --no-cache git gcc musl-dev

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the Go application
RUN go build -o carnac .

# Final stage
FROM alpine:latest

# Install necessary runtime dependencies
RUN apk --no-cache add ca-certificates mysql-client

# Create a non-root user to run the application
RUN addgroup -g 1000 carnac && \
    adduser -D -u 1000 -G carnac carnac

# Set the working directory inside the container
WORKDIR /app

# Copy the built application from the builder stage
COPY --from=builder /app/carnac .

# Switch to the non-root user
USER carnac

# Expose the application port
EXPOSE 8080

# Environment variables
ENV DBUSER="root" \
    DBPASSWORD="" \
    DBHOST="localhost" \
    DBPORT="3306" \
    DBNAME="carnac" \
    LOGLEVEL="info"

# Health check to ensure the application is running
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/status || exit 1

# Run the application in server mode
CMD ["./carnac", "--server", "--port", "8080"]