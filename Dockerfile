# ✅ Используем актуальную и безопасную версию Go
FROM golang:1.24.3-alpine

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Устанавливаем зависимости для Alpine (git нужен для go get, если используешь)
RUN apk add --no-cache git

# Копируем файлы зависимостей
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем весь исходный код
COPY . .

# Собираем приложение
RUN go build -o main ./cmd/marketflow  

# Открываем порт
EXPOSE 8080

# Команда по умолчанию для запуска
CMD ["./main"]
