FROM golang:latest AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /cdd-go ./cmd/cdd-go

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
COPY --from=builder /cdd-go /cdd-go

ENTRYPOINT ["/cdd-go", "serve_json_rpc"]
CMD ["--port", "8082", "--listen", "0.0.0.0"]
