# Order Monitoring API Sentry Demo

Учебный проект на Go/Gin для демонстрации мониторинга, трассировки ошибок и инцидентов через Sentry.

В репозитории есть две версии:

```txt
broken-version/  — намеренно сломанная версия для демонстрации ошибок и Sentry issues
fixed-version/   — исправленная версия с валидацией, защитой от гонки в оплате и нормальной обработкой ошибок
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

Для broken-version:

```bash
cd broken-version
cp .env.example .env
# вставьте SENTRY_DSN в .env, если хотите отправлять события в Sentry

docker compose up --build
```

Для fixed-version:

```bash
cd fixed-version
cp .env.example .env
# вставьте SENTRY_DSN в .env, если хотите отправлять события в Sentry

docker compose up --build
```

## Быстрая проверка

```bash
curl http://localhost:8080/health
curl http://localhost:8081/health
```

## Что демонстрирует проект

1. Panic при расчете скидки с `discount_percent = 0` в `broken-version`.
2. Ошибку базы данных при повторном создании заказа без уникального `order_number` в `broken-version`.
3. Медленный запрос для демонстрации tracing/performance в `broken-version`.
4. Race condition в оплате: под конкурентной нагрузкой один заказ может быть оплачен несколько раз в `broken-version`.
5. Исправленное поведение тех же сценариев в `fixed-version`.

Подробности по сценариям находятся в [broken-version/README.md](broken-version/README.md) и [fixed-version/README.md](fixed-version/README.md).

Для Postman в корне лежит `order-monitoring.postman_collection.json`.
