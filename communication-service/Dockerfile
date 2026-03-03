ARG GO_VERSION=1.26

# ====== Build stage ======
FROM golang:${GO_VERSION}-alpine AS builder

WORKDIR /app

# Копируем shared модуль
COPY shared/go.mod shared/go.sum ./shared/
COPY shared/ ./shared/

# Копируем сервис
COPY communication-service/go.mod communication-service/go.sum ./communication-service/
WORKDIR /app/communication-service
RUN go mod download

WORKDIR /app
COPY communication-service/ ./communication-service/

WORKDIR /app/communication-service
RUN CGO_ENABLED=0 go build -o /communication-service ./cmd/api

# ====== Run stage ======
FROM alpine:latest

WORKDIR /root/

COPY --from=builder /communication-service .
COPY --from=builder /app/communication-service/config ./config

CMD ["./communication-service"]
