ARG GO_VERSION=1.26

# ====== Build stage ======
FROM golang:${GO_VERSION}-alpine AS builder

WORKDIR /app

# Копируем shared модуль
COPY shared/go.mod shared/go.sum ./shared/
COPY shared/ ./shared/

# Копируем сервис
COPY api-gateway/go.mod api-gateway/go.sum ./api-gateway/
WORKDIR /app/api-gateway
RUN go mod download

WORKDIR /app
COPY api-gateway/ ./api-gateway/

WORKDIR /app/api-gateway
RUN CGO_ENABLED=0 go build -o /api-gateway ./cmd/api

# ====== Run stage ======
FROM alpine:latest

WORKDIR /root/

COPY --from=builder /api-gateway .
COPY --from=builder /app/api-gateway/config ./config

CMD ["./api-gateway"]
