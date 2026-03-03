FROM golang:alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /cdd-go ./cmd/cdd-go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /cdd-go /cdd-go

ENTRYPOINT ["/cdd-go", "serve_json_rpc"]
CMD ["--port", "8082", "--listen", "0.0.0.0"]
