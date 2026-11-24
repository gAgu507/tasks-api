# Стадия 1: сборка сервера
FROM golang:1.27 AS build
WORKDIR /src

# Сначала только зависимости: этот слой кэшируется и не пересобирается,
# пока go.mod и go.sum не изменились
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /out/server ./cmd/server

# Стадия 2: утилита goose
FROM golang:1 AS goose
RUN CGO_ENABLED=0 go install github.com/pressly/goose/v3/cmd/goose@latest

# Стадия 3: образ для миграций
FROM alpine:3 AS migrate
COPY --from=goose /go/bin/goose /goose
COPY migrations /migrations
ENTRYPOINT ["/goose"]

# Стадия 2: итоговый образ
FROM alpine:3
COPY --from=build /out/server /server
EXPOSE 8080
ENTRYPOINT ["/server"]