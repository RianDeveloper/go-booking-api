# Go Booking API

## Стек

- Go
- net/http
- PostgreSQL
- Kafka
- Docker Compose
- SQL migrations

## Запуск

1. Скопировать `.env.example` в `.env`
2. Поднять инфраструктуру:

```bash
make up
```

3. Применить миграции:

```bash
make migrate-up
```

4. Запустить API:

```bash
make run
```

API будет доступен на `http://localhost:8080`.

## Основные эндпоинты

- `POST /v1/users`
- `POST /v1/resources`
- `POST /v1/resources/{id}/slots`
- `GET /v1/resources/{id}/schedule`
- `POST /v1/bookings`
- `GET /v1/bookings/{id}`
- `POST /v1/bookings/{id}/cancel`

## Пример сценария

1. Создай пользователя:
   - `POST /v1/users` с `email`, `name`
2. Создай ресурс:
   - `POST /v1/resources` с `name`
3. Создай слоты для ресурса:
   - `POST /v1/resources/{id}/slots` с массивом `start_time`, `end_time`
4. Создай бронь:
   - `POST /v1/bookings` с `user_id`, `slot_id`
5. Получи бронь:
   - `GET /v1/bookings/{id}`
6. Отмени бронь:
   - `POST /v1/bookings/{id}/cancel`
7. Проверь расписание ресурса:
   - `GET /v1/resources/{id}/schedule?from=<RFC3339>&to=<RFC3339>`

## Тесты

Unit:
- `internal/service/booking_service_test.go`

HTTP:
- `internal/transport/http/handler_test.go`

Integration:
- `internal/repository/postgres/store_integration_test.go`

Запуск:

```bash
make test
```

Проверка гонок:

```bash
make race
```

## Архитектура

- `cmd/api` — entrypoint
- `internal/config` — конфиг через env
- `internal/domain` — сущности и ошибки
- `internal/ports` — интерфейсы доступа к данным
- `internal/service` — бизнес-логика
- `internal/repository/postgres` — PostgreSQL и транзакции
- `internal/outbox` — outbox worker и Kafka публикация
- `internal/transport/http` — REST handlers и ошибки
- `migrations` — SQL миграции
