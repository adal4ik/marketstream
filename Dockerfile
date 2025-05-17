# Используем официальный образ Golang в качестве базового
FROM golang:1.23.2-alpine

# Устанавливаем рабочую директорию
WORKDIR /app

# Устанавливаем зависимости для Alpine
RUN apk add --no-cache git

# Копируем go.mod и go.sum
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN go build -o main ./cmd/app

# Открываем порт, на котором работает приложение
EXPOSE 8080

# Команда для запуска приложения
CMD ["./main"]