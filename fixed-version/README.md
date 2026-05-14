# Fixed Version — Order Monitoring API

Эта версия исправляет ошибки из `broken-version`:

- division by zero заменен на валидацию `discount_percent`;
- duplicate order number больше не возникает автоматически;
- slow request отключен;
- гонка при оплате исправлена транзакцией и блокировкой строки заказа;
- клиентские ошибки возвращают нормальные HTTP-коды;
- неожиданные серверные ошибки отправляются в Sentry вручную.

## Запуск

```bash
cp .env.example .env
# вставьте SENTRY_DSN в .env, если нужен реальный Sentry
# для Elastic включите ELASTIC_ENABLED=true и заполните ELASTIC_* переменные
docker compose up --build
```

## Логирование и Elastic

- Логи пишутся в JSON (`slog`) в stdout.
- При `ELASTIC_ENABLED=true` те же логи отправляются в Elasticsearch (`elastic.co`).
- Для подключения укажите `ELASTIC_CLOUD_ID` + `ELASTIC_API_KEY` (или `ELASTIC_URL` + креды).

Проверка:

```bash
curl http://localhost:8081/health
curl http://localhost:8081/products
```

Swagger UI:

```txt
http://localhost:8081/swagger/index.html
```

OpenAPI JSON:

```txt
http://localhost:8081/swagger/doc.json
```

## Проверка исправления 1: discount validation

```bash
curl -X POST http://localhost:8081/orders/discount           -H "Content-Type: application/json"           -d '{"price":1000,"discount_percent":0}'
```

Ожидаемый результат:

```json
{"error":"discount_percent must be between 1 and 100"}
```

Сервер не падает.

Корректный запрос:

```bash
curl -X POST http://localhost:8081/orders/discount           -H "Content-Type: application/json"           -d '{"price":1000,"discount_percent":10}'
```

## Проверка исправления 2: order number

Теперь, если `order_number` не передан, backend генерирует уникальный номер:

```bash
curl -X POST http://localhost:8081/orders           -H "Content-Type: application/json"           -d '{"product_id":1,"quantity":1,"customer_email":"student@example.com"}'
```

Повторный запрос тоже должен пройти успешно.

## Проверка исправления 3: slow request disabled

```bash
curl -X POST http://localhost:8081/orders           -H "Content-Type: application/json"           -d '{"product_id":1,"quantity":1,"customer_email":"slow@example.com","simulate_slow":true}'
```

Ожидаемый результат:

```json
{"error":"simulate_slow is disabled in fixed version"}
```

## Проверка исправления 4: race condition в оплате

Сначала создайте заказ:

```bash
curl -X POST http://localhost:8081/orders \
  -H "Content-Type: application/json" \
  -d '{"product_id":1,"quantity":1,"customer_email":"race@example.com"}'
```

Потом одновременно отправьте несколько запросов на оплату того же заказа:

```bash
ORDER_ID=1
seq 1 20 | xargs -I{} -P 20 curl -s -X POST http://localhost:8081/orders/$ORDER_ID/payments
```

Проверка результата:

```bash
curl http://localhost:8081/orders/1
curl http://localhost:8081/orders/1/payments
```

Ожидаемый результат:

- только один запрос создает платеж;
- остальные получают контролируемый ответ `409 order already paid`;
- `paid_amount` равен `total_price`, а не превышает его;
- в списке платежей остается одна запись.

## Что показывать на защите

1. Те же запросы, что ломали broken version.
2. API больше не падает.
3. Ошибка превращена в контролируемый HTTP-ответ.
4. Код стал надежнее за счет валидации, обработки ошибок и блокировки строки заказа при оплате.
