ARG AUTH_SERVICE_PORT
ARG GO_VERSION

# ====== Build stage ======
FROM golang:${GO_VERSION}-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod tidy

COPY . .

RUN go build -o auth-service .

# ====== Run stage ======
FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/auth-service .

EXPOSE ${AUTH_SERVICE_PORT}

CMD ["./auth-service"]
