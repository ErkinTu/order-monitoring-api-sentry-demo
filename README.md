# Order Monitoring API Sentry Demo

Учебный проект на Go/Gin для демонстрации мониторинга, трассировки ошибок и инцидентов через Sentry.

Сейчас в репозиторий добавлена только версия:

```txt
broken-version/  — намеренно сломанная версия для демонстрации ошибок и Sentry issues
```

## Стек

- Go
- Gin
- GORM
- PostgreSQL
- Docker Compose
- Sentry Go SDK
- Sentry Gin middleware

## Запуск

```bash
cd broken-version
cp .env.example .env
# вставьте SENTRY_DSN в .env, если хотите отправлять события в Sentry

docker compose up --build
```

## Быстрая проверка

```bash
curl http://localhost:8080/health
curl http://localhost:8080/products
```

## Что демонстрирует broken-version

1. Panic при расчете скидки с `discount_percent = 0`.
2. Ошибку базы данных при повторном создании заказа без уникального `order_number`.
3. Медленный запрос для демонстрации tracing/performance.

Подробности по сценариям находятся в [broken-version/README.md](broken-version/README.md).
