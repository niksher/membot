# Этап сборки
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Копируем файлы модулей
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o tg-video-bot ./cmd/bot/

# Этап запуска
FROM alpine:3.18

WORKDIR /app

# Устанавливаем зависимости для MySQL
RUN apk add --no-cache ca-certificates tzdata

# Копируем бинарник и конфиги
COPY --from=builder /app/tg-video-bot .
COPY --from=builder /app/.env .

# Создаем пользователя для безопасности
RUN adduser -D -g '' appuser
USER appuser

CMD ["./tg-video-bot"]