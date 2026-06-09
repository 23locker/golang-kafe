# ---- Build stage ----
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Копируем зависимости отдельно для кеширования слоев
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь исходный код
COPY . .

# Собираем бинарник (CGO отключен для чистого статического бинарника)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/server ./cmd/app

# ---- Runtime stage ----
FROM alpine:3.19

# Нужен для HTTPS-запросов и корректного времени
RUN apk --no-cache add ca-certificates tzdata

# Устанавливаем временную зону (Байкал / Иркутск)
ENV TZ=Asia/Irkutsk

WORKDIR /app

# Копируем бинарник из builder-стадии
COPY --from=builder /app/server .

# Копируем SQL-миграции (читаются при старте из main.go)
COPY --from=builder /app/migrations ./migrations

# Директория для загружаемых файлов (фото блюд)
RUN mkdir -p /app/uploads

EXPOSE 8080

CMD ["./server"]
