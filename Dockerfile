FROM golang:1.25-alpine AS builder
LABEL authors="xzescha"
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server ./cmd/server

FROM alpine:3.20

WORKDIR /app

RUN adduser -D -g '' appuser
USER appuser

COPY --from=builder /app/server /app/server

EXPOSE ${HTTP_PORT}

ENTRYPOINT ["/app/server"]