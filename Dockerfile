FROM golang:1.22.6-alpine AS builder

# Устанавливаем рабочую директорию
WORKDIR /app

# Устанавливаем необходимые зависимости для сборки
RUN apk update && apk add --no-cache git

# Копируем go.mod и go.sum из папки проекта
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код из папки проекта
COPY . .

# Собираем исполняемый файл
RUN go build -o main cmd/app/main.go

# Используем минимальный образ Alpine для запуска
FROM alpine:3.20

# Устанавливаем необходимые зависимости для работы приложения
RUN apk add --no-cache libgcc curl

# Копируем собранный исполняемый файл из предыдущего этапа
COPY --from=builder /app/main /main

# Копируем .env файл из предыдущего этапа
COPY --from=builder /app/.env /app/.env

# Устанавливаем рабочую директорию
WORKDIR /app

# Проверяем содержимое директории /app
RUN ls -la /app

# Задаем команду по умолчанию
ENTRYPOINT ["/main"]

# Открываем порт, если приложение слушает на порту (например, 8080)
EXPOSE 8080
