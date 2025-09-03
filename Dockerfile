# Go official image as a build stage
FROM golang:1.24.5-alpine AS build

# working directory inside the container
WORKDIR /app

# Install git for fetching dependencies
RUN apk add --no-cache git

# Copy go mod and sum files and cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Use a minimal image for the final stage
FROM alpine:latest

# Install necessary certificates and timezone data
RUN apk --no-cache add ca-certificates tzdata

# Create a non-root user to run the application
RUN adduser -D -s /bin/sh appuser

# Set the working directory
WORKDIR /root/

# Copy the binary from the build stage
COPY --from=build /app/main .

# Ensure the binary is executable and owned by the non-root user
RUN chown appuser:appuser main

# Switch to the non-root user
USER appuser

# Expose the application port for future use
EXPOSE 8080

# Set environment variables for database connection
ENV DBUSER=root
ENV DBPASSWORD=password
ENV DBHOST=localhost
ENV DBPORT=3306
ENV DBNAME=carnac

# Command to run the application
CMD ["./main"]