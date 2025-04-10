# Dockerfile (в папке content-generator/)
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o content-generator ./cmd/server/main.go

# Финальный образ
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/content-generator .
COPY .env .

EXPOSE 8080
CMD ["./content-generator"]
