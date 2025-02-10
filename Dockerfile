## build
FROM golang:1.23.3-alpine3.20 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o insider ./cmd/main.go

# runtime
FROM alpine:3.20.3

WORKDIR /

COPY --from=builder /app/insider /insider
COPY --from=builder /app/internal/business/domain/message/stores/db/postgres/migrations /migrations 

EXPOSE 8080
ENTRYPOINT ["/insider"]
