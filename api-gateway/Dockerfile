# ====== Build stage ======
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Скопируем go.mod и go.sum для установки зависимостей
COPY go.mod go.sum ./
RUN go mod tidy

# Копируем исходники
COPY . .

# Собираем бинарник
RUN go build -o api-gateway .

# ====== Run stage ======
FROM alpine:latest

WORKDIR /root/

# Копируем бинарник из builder
COPY --from=builder /app/api-gateway .

EXPOSE 8080

CMD ["./api-gateway"]
