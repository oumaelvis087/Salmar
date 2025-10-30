FROM golang:1.20-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/salmar-api ./cmd/api

# Use a small alpine image for the final container
FROM alpine:3.16

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/salmar-api .

# Expose the port
EXPOSE 8080

# Run the application
CMD ["/app/salmar-api"]