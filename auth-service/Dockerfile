ARG GO_VERSION=1.26

# ====== Build stage ======
FROM golang:${GO_VERSION}-alpine AS builder

WORKDIR /app

# Копируем shared модуль
COPY shared/go.mod shared/go.sum ./shared/
COPY shared/ ./shared/

# Копируем сервис
COPY auth-service/go.mod auth-service/go.sum ./auth-service/
WORKDIR /app/auth-service
RUN go mod download

WORKDIR /app
COPY auth-service/ ./auth-service/

WORKDIR /app/auth-service
RUN CGO_ENABLED=0 go build -o /auth-service ./cmd/api

# ====== Run stage ======
FROM alpine:latest

WORKDIR /root/

COPY --from=builder /auth-service .
COPY --from=builder /app/auth-service/config ./config

CMD ["./auth-service"]
