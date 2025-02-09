# Этап сборки
FROM golang:1.20 AS builder

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем go.mod и go.sum, чтобы установить зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем остальные файлы приложения
COPY . .

# Компилируем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o manager .

# Этап выполнения
FROM alpine:latest

# Устанавливаем необходимые библиотеки
RUN apk --no-cache add ca-certificates

# Копируем скомпилированное приложение из этапа сборки
COPY --from=builder /app/manager /manager

# Устанавливаем переменные окружения
ENV MANAGER_PORT=18081

# Открываем порт для приложения
EXPOSE 18081

# Запускаем приложение
CMD ["/manager"]
