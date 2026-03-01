# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the Go application statically
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o cdd-go ./cmd/cdd-go

# Final stage
FROM alpine:latest

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/cdd-go .

# Set the entrypoint
ENTRYPOINT ["./cdd-go", "server_json_rpc"]
