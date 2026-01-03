# Build stage
FROM golang:alpine AS builder

WORKDIR /app

# Install dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build the combined engine binary
RUN CGO_ENABLED=0 GOOS=linux go build -o engram-engine ./cmd/combined/main.go

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/engram-engine .

# Default environment variables
ENV STORE_TYPE=badger
ENV DB_PATH=/app/data
ENV GRPC_PORT=50051
ENV REDIS_URL=redis:6379

# Create data directory for Badger
RUN mkdir -p /app/data

EXPOSE 50051

CMD ["./engram-engine"]
