# Tasks API

Простой REST API для списка задач на Go с хранением данных в PostgreSQL. Учебный проект: CRUD, работа с базой через `pgx`, миграции, тесты обработчиков, обработка ошибок и коды ответов HTTP.

## Что нужно

- Go 1.22 или новее (в проекте используется маршрутизация по методам и параметры пути из стандартной библиотеки)
- PostgreSQL
- [goose](https://github.com/pressly/goose) для миграций базы:

  ```bash
  go install github.com/pressly/goose/v3/cmd/goose@latest
  ```

  Если команда `goose` не находится, добавь папку `go/bin` в `PATH` (`~/go/bin` на Linux, `%USERPROFILE%\go\bin` на Windows).

## Запуск

1. Клонировать репозиторий и перейти в папку проекта:

   ```bash
   git clone https://github.com/gAgu507/tasks-api.git
   cd tasks-api
   ```

2. Создать базу данных:

   ```bash
   createdb tasks
   ```

   Если `createdb` не находит пользователя, создай базу через `psql` от имени суперпользователя (на Linux обычно `sudo -u postgres psql`).

3. Задать настройки через переменные окружения (см. таблицу ниже).

   Linux / macOS / Git Bash:

   ```bash
   export DATABASE_URL="postgres://пользователь:пароль@localhost:5432/tasks"
   ```

   PowerShell:

   ```powershell
   $env:DATABASE_URL = "postgres://пользователь:пароль@localhost:5432/tasks"
   ```

4. Применить миграции (создаёт структуру таблиц):

   ```bash
   goose -dir migrations postgres "$DATABASE_URL" up
   ```

   В PowerShell: `goose -dir migrations postgres $env:DATABASE_URL up`.

5. Запустить сервер:

   ```bash
   go run ./cmd/server
   ```

   По умолчанию сервер слушает `http://localhost:8080`.

## Настройки

| Переменная | Обязательна | По умолчанию | Описание |
|------------|-------------|--------------|----------|
| `DATABASE_URL` | да | | Строка подключения к PostgreSQL |
| `PORT` | нет | `8080` | Порт, на котором слушает сервер |

Без `DATABASE_URL` сервер сообщает об этом и завершает работу.

Пример с другим портом:

```bash
PORT=9090 go run ./cmd/server
```

В PowerShell переменная задаётся отдельной строкой: `$env:PORT = "9090"`.

## Миграции

Структура базы описана миграциями в папке `migrations/`: это пронумерованные SQL-файлы, каждый с одним изменением. Утилита goose помнит, какие из них уже применены, и накатывает только новые.

```bash
goose -dir migrations postgres "$DATABASE_URL" status   # что применено
goose -dir migrations postgres "$DATABASE_URL" up       # применить новые
goose -dir migrations postgres "$DATABASE_URL" down     # откатить последнюю
```

Новая миграция создаётся командой:

```bash
goose -dir migrations create название_изменения sql
```

Уже применённые миграции не редактируются: исправление оформляется новой миграцией.

## Тесты

```bash
go test ./...
```

Тесты обработчиков работают без базы данных и сети: вместо PostgreSQL подставляется фальшивое хранилище в памяти, которое реализует интерфейс `TaskStore`. Подробный вывод по каждому случаю:

```bash
go test ./internal/handlers -v
```

Покрытие кода:

```bash
go test -cover ./...
```

## Адреса

| Метод | Путь | Описание | Коды ответа |
|-------|------|----------|-------------|
| GET | `/` | Проверка, что сервер работает | 200 |
| GET | `/tasks` | Список всех задач | 200, 500 |
| POST | `/tasks` | Создать задачу | 201, 400, 500 |
| GET | `/tasks/{id}` | Получить задачу по id | 200, 400, 404, 500 |
| PUT | `/tasks/{id}` | Заменить задачу целиком | 204, 400, 404, 500 |
| DELETE | `/tasks/{id}` | Удалить задачу | 204, 400, 404, 500 |

Значения кодов: `400`: неверный id или некорректный JSON, `404`: задачи с таким id нет, `500`: внутренняя ошибка (например, недоступна база).

## Формат задачи

```json
{
  "id": 1,
  "title": "Купить хлеб",
  "description": "Ржаной, в магазине у дома",
  "completed": false
}
```

При создании обязателен только `title`. Поля `id` и `completed` назначает сервер.

`PUT` заменяет задачу целиком: если не передать `completed`, оно станет `false`.

## Примеры

Создать задачу:

```bash
curl -X POST http://localhost:8080/tasks \
  -H "Content-Type: application/json" \
  -d '{"title": "Купить хлеб", "description": "Ржаной"}'
```

Получить список:

```bash
curl http://localhost:8080/tasks
```

Отметить задачу выполненной:

```bash
curl -X PUT http://localhost:8080/tasks/1 \
  -H "Content-Type: application/json" \
  -d '{"title": "Купить хлеб", "description": "Ржаной", "completed": true}'
```

Удалить задачу:

```bash
curl -X DELETE http://localhost:8080/tasks/1
```

## Структура проекта

```
cmd/server/main.go                точка входа: настройки, подключение к базе, запуск
internal/config/config.go         чтение переменных окружения
internal/store/store.go           структура Task, ошибка ErrNotFound, интерфейс TaskStore
internal/store/postgres.go        реализация хранилища на PostgreSQL
internal/handlers/handlers.go     обработчики HTTP-запросов и маршруты
internal/handlers/response.go     вспомогательные функции для JSON-ответов
internal/handlers/*_test.go       тесты обработчиков и фальшивое хранилище
migrations/                       миграции базы данных (goose)
```

## Планы

- Docker и `docker compose`
- Авторизация (регистрация, вход, JWT)
- Фильтры и постраничный вывод
- Логирование через `slog`, корректная остановка сервера
- CI на GitHub Actions
