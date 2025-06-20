# Stage 1: Build and Test
FROM golang:1.24.4-alpine3.22 AS builder

WORKDIR /app

# Install git for go mod and any dependencies
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Run tests
RUN go test ./...

# Build the binary
RUN go build -o postmark-exporter .

# Stage 2: Run
FROM alpine:3.22 AS runtime

WORKDIR /app

# Copy the built binary from builder
COPY --from=builder /app/postmark-exporter .

# Expose port (change if needed)
EXPOSE 8080

# Run the application
ENTRYPOINT ["./postmark-exporter"]