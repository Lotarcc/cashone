# Build stage
FROM golang:1.20-alpine AS builder

WORKDIR /app

# Install git and build dependencies
RUN apk add --no-cache git

# Copy go mod and sum files
COPY src/app/go.mod src/app/go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY src/app/ .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Final stage
FROM alpine:3.17

WORKDIR /app

# Install certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Copy the binary from builder
COPY --from=builder /app/main .

# Expose port
EXPOSE 8081

# Run the application
CMD ["./main"]
