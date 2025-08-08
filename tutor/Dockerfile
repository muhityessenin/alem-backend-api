# Dockerfile

# 1. Этап сборки (Builder)
# Укажите вашу версию Go.
FROM golang:1.23-alpine AS builder
# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем файлы go.mod и go.sum для кэширования зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь исходный код сервиса
COPY . .

# Собираем приложение Go. Бинарный файл будет называться 'server'
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/main.go

# 2. Финальный образ (Final)
# Используем самый минимальный и безопасный образ Alpine
FROM alpine:latest

WORKDIR /app

# Копируем только скомпилированный бинарный файл из этапа сборки
COPY --from=builder /app/server .

# ВАЖНО: Укажите порт, который слушает ваше приложение
EXPOSE 8080

# Команда для запуска приложения
CMD ["./server"]