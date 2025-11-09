# syntax=docker/dockerfile:1

# Use an official Go image as the builder
FROM golang:1.21-alpine AS builder

# Install necessary packages
RUN apk add --no-cache tzdata

# Set the timezone
ENV TZ=Africa/Nairobi
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# Set the working directory
WORKDIR /app

# Cache dependencies by copying go.mod and go.sum first
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . ./

# Build the Go application
RUN go build -o /shortcode-service

# Start a new stage with a minimal image for the final build
FROM alpine:3.18

# Install tzdata
RUN apk add --no-cache tzdata

# Set the timezone
ENV TZ=Africa/Nairobi
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# Copy everything from the build stage to ensure migrations and other necessary files are included
COPY --from=builder /app /app
COPY --from=builder /shortcode-service /shortcode-service

# Expose the HTTP port
EXPOSE 80

# Command to run the application
CMD ["/shortcode-service"]
