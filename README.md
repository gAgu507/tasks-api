# Tasks API

Простой REST API для списка задач на Go с хранением данных в PostgreSQL. Учебный проект: CRUD, работа с базой через `pgx`, миграции, тесты обработчиков, Docker Compose, обработка ошибок и коды ответов HTTP.

## Быстрый старт через Docker Compose

Самый простой способ: Go и PostgreSQL ставить не нужно, достаточно Docker (на Linux можно Podman).

1. Клонировать репозиторий и перейти в папку проекта:

   ```bash
   git clone https://github.com/gAgu507/tasks-api.git
   cd tasks-api
   ```

2. Создать файл настроек из шаблона:

   Linux / macOS / Git Bash:

   ```bash
   cp .env.example .env
   ```

   PowerShell:

   ```powershell
   copy .env.example .env
   ```

   В `.env` можно оставить значения по умолчанию или задать свои `POSTGRES_USER`, `POSTGRES_PASSWORD` и `POSTGRES_DB`. Пароль для Compose лучше делать без спецсимволов.

3. Запустить всё одной командой:

   ```bash
   docker compose up --build
   ```

   На Linux с Podman: `podman compose up --build` (при необходимости `sudo dnf install podman-compose`).

   Compose поднимет три контейнера по порядку: PostgreSQL, затем миграции, затем сервер. Первый запуск дольше остальных, потому что скачиваются и собираются образы.

4. Проверить, что сервер работает:

   ```bash
   curl http://localhost:8080/tasks
   ```

   Ответ на чистой базе: `[]`.

Полезные команды:

```bash
docker compose up -d --build   # запуск в фоне
docker compose logs -f app     # логи сервера
docker compose down            # остановить и удалить контейнеры (данные сохраняются)
docker compose down -v         # то же, плюс удалить данные базы
```

База дополнительно доступна с твоего компьютера на порту `5433` (например, для `psql`). Данные хранятся в томе `pgdata` и переживают перезапуск контейнеров.

## Запуск без Docker

### Что нужно

- Go 1.22 или новее (в проекте используется маршрутизация по методам и параметры пути из стандартной библиотеки)
- PostgreSQL
- [goose](https://github.com/pressly/goose) версии 3.24.0 или новее для миграций базы (эта версия умеет читать настройки из `.env`):

  ```bash
  go install github.com/pressly/goose/v3/cmd/goose@latest
  ```

  Если команда `goose` не находится, добавь папку `go/bin` в `PATH` (`~/go/bin` на Linux, `%USERPROFILE%\go\bin` на Windows).

### Шаги

1. Клонировать репозиторий (см. выше) и создать базу данных:

   ```bash
   createdb tasks
   ```

   Если `createdb` не находит пользователя, создай базу через `psql` от имени суперпользователя (на Linux обычно `sudo -u postgres psql`).

2. Создать `.env` из шаблона и подставить свои значения для локальной базы (`DATABASE_URL` и `GOOSE_DBSTRING` с адресом `localhost`).

3. Применить миграции:

   ```bash
   goose up
   ```

4. Запустить сервер:

   ```bash
   go run ./cmd/server
   ```

   По умолчанию сервер слушает `http://localhost:8080`.

## Настройки

Настройки читаются из файла `.env` в корне проекта. Шаблон лежит в `.env.example`. Сервер и `goose` нужно запускать из корня проекта, где лежит этот файл.

| Переменная | Для чего | По умолчанию | Описание |
|------------|----------|--------------|----------|
| `POSTGRES_USER` | Compose | | Пользователь базы в контейнере |
| `POSTGRES_PASSWORD` | Compose | | Пароль пользователя |
| `POSTGRES_DB` | Compose | | Имя базы |
| `DATABASE_URL` | локальный запуск | | Строка подключения к PostgreSQL для сервера (обязательна) |
| `PORT` | локальный запуск | `8080` | Порт, на котором слушает сервер |
| `GOOSE_DRIVER` | локальные миграции | | Драйвер базы для goose, здесь `postgres` |
| `GOOSE_DBSTRING` | локальные миграции | | Строка подключения к базе для goose |
| `GOOSE_MIGRATION_DIR` | локальные миграции | | Папка с миграциями, здесь `./migrations` |

В Docker Compose адрес базы и порт собираются из переменных `POSTGRES_*` автоматически, поэтому `DATABASE_URL` и `GOOSE_*` нужны только для запуска без Docker.

Пример `.env`:

```
POSTGRES_USER=tasks_user
POSTGRES_PASSWORD=tasks_pass
POSTGRES_DB=tasks

DATABASE_URL=postgres://tasks_user:tasks_pass@localhost:5433/tasks
PORT=8080

GOOSE_DRIVER=postgres
GOOSE_DBSTRING=postgres://tasks_user:tasks_pass@localhost:5433/tasks
GOOSE_MIGRATION_DIR=./migrations
```

В этом примере локальный запуск подключается к базе из Compose на порт `5433`. Для обычного локального PostgreSQL укажи свои пользователя, пароль и порт `5432`.

Правила записи: без пробелов вокруг `=` и без кавычек. Спецсимволы в пароле (`@`, `:`, `/`, `#`) нужно кодировать в адресе, например `@` как `%40`.

Переменные, заданные в окружении (например, через `export` или в Docker), главнее значений из `.env`. Без `DATABASE_URL` сервер сообщает об этом и завершает работу.

Пример с другим портом без правки файла:

```bash
PORT=9090 go run ./cmd/server
```

В PowerShell переменная задаётся отдельной строкой: `$env:PORT = "9090"`.

## Миграции

Структура базы описана миграциями в папке `migrations/`: это пронумерованные SQL-файлы, каждый с одним изменением. Утилита goose помнит, какие из них уже применены, и накатывает только новые. В Docker Compose миграции применяются автоматически при каждом запуске.

При локальном запуске настройки подключения goose берёт из `.env`:

```bash
goose status   # что применено
goose up       # применить новые
goose down     # откатить последнюю
```

Новая миграция создаётся командой:

```bash
goose create название_изменения sql
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

В PowerShell вместо `curl` пиши `curl.exe` и экранируй кавычки в JSON (`\"`), либо используй `Invoke-RestMethod`:

```powershell
Invoke-RestMethod -Method Post -Uri http://localhost:8080/tasks `
  -ContentType "application/json" `
  -Body '{"title": "Купить хлеб", "description": "Ржаной"}'
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
Dockerfile                        сборка сервера и образа для миграций
compose.yaml                      запуск базы, миграций и сервера вместе
.env.example                      шаблон настроек (сам .env в Git не попадает)
```

## Планы

- Авторизация (регистрация, вход, JWT)
- Фильтры и постраничный вывод
- Логирование через `slog`, корректная остановка сервера
- CI на GitHub Actions
