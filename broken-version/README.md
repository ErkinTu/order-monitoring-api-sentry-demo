# Broken Version — Order Monitoring API

Эта версия намеренно содержит ошибки для демонстрации Sentry.

## Запуск

```bash
cp .env.example .env
# вставьте SENTRY_DSN в .env, если нужен реальный Sentry
docker compose up --build
```

Проверка:

```bash
curl http://localhost:8080/health
curl http://localhost:8080/products
```

## Ошибка 1: panic / division by zero

```bash
curl -X POST http://localhost:8080/orders/discount           -H "Content-Type: application/json"           -d '{"price":1000,"discount_percent":0}'
```

Ожидаемый результат:

- backend получает panic: integer divide by zero;
- Sentry создает issue;
- в stack trace видно место ошибки в `internal/order/service.go`.

## Ошибка 2: duplicate key в PostgreSQL

Первый запрос создаст заказ:

```bash
curl -X POST http://localhost:8080/orders           -H "Content-Type: application/json"           -d '{"product_id":1,"quantity":1,"customer_email":"student@example.com"}'
```

Второй такой же запрос упадет на unique constraint, потому что в broken version всем заказам без `order_number` ставится одинаковый номер:

```bash
curl -X POST http://localhost:8080/orders           -H "Content-Type: application/json"           -d '{"product_id":1,"quantity":1,"customer_email":"student@example.com"}'
```

## Ошибка 3: slow request

```bash
curl -X POST http://localhost:8080/orders           -H "Content-Type: application/json"           -d '{"product_id":1,"quantity":1,"customer_email":"slow@example.com","order_number":"SLOW-001","simulate_slow":true}'
```

Ожидаемый результат:

- запрос выполняется примерно 3 секунды;
- при включенном tracing Sentry может показать медленную транзакцию.

## Что показывать на защите

1. Запрос, который вызывает ошибку.
2. Sentry issue.
3. Stack trace.
4. Файл и строку, где ошибка возникла.
5. Переход к `fixed-version` и повторную проверку.
