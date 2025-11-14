FROM golang:1.25-alpine AS builder
LABEL authors="xzescha"
WORKDIR /app

# зависимости отдельно — чтобы кешировалось
COPY go.mod go.sum ./
RUN go mod download

# остальной код
COPY . .

# собираем бинарь
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server ./cmd/server

# --- runtime stage ---
FROM alpine:3.20

WORKDIR /app

# не обязателен, но приятно
RUN adduser -D -g '' appuser
USER appuser

COPY --from=builder /app/server /app/server

EXPOSE 8080

# переменные по умолчанию (можно перекрыть в docker-compose)
ENV HTTP_PORT=8080

ENTRYPOINT ["/app/server"]