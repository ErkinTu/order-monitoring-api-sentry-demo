# Postman scenarios for monitoring demo

## Подготовка

1. Импортируйте в Postman файл `order-monitoring.postman_collection.json`.
2. Запустите `broken-version` на `http://localhost:8080`.
3. Запустите `fixed-version` на `http://localhost:8081`.
4. В Collection Variables можно менять:
   - `client_iterations` - сколько раз имитировать клиентский цикл;
   - `client_delay_ms` - пауза между клиентскими циклами;
   - `concurrent_payment_requests` - сколько одновременных оплат отправить для race condition.

Для короткой презентации оставьте значения по умолчанию:

```txt
client_iterations = 8
client_delay_ms = 600
concurrent_payment_requests = 12
```

## Сценарий 1: broken-version

В Postman откройте Collection Runner и запустите папку:

```txt
Presentation Runner · Broken Load
```

Что делает сценарий:

1. Проверяет `/health`.
2. Несколько раз имитирует клиента:
   - `GET /products`
   - `POST /orders`
   - `GET /orders/:id`
   - `POST /orders/discount`
3. После обычного потока специально показывает ошибки:
   - duplicate order number -> `500`;
   - discount with zero percent -> panic / `500`;
   - slow request -> заметная задержка около 3 секунд.

Во время показа откройте Elastic Discover и фильтруйте:

```txt
service.name : "order-monitoring-api-broken"
```

Полезные поля для таблицы:

```txt
@timestamp
service.name
http.method
http.path
http.status_code
duration_ms
error
```

Что визуально показать:

- рост количества `500`;
- медленный запрос с большим `duration_ms`;
- ошибки duplicate key и division by zero;
- в Sentry можно открыть issue/stack trace для panic.

## Сценарий 2: fixed-version

После broken-сценария запустите папку:

```txt
Presentation Runner · Fixed Load
```

Это тот же пользовательский поток, но на:

```txt
http://localhost:8081
```

Ожидаемая разница:

- duplicate order number возвращает `409`, а не `500`;
- discount with zero percent возвращает `400`, а не panic;
- `simulate_slow` возвращает `400` быстро, без задержки;
- обычный клиентский поток продолжает работать.

В Elastic переключите фильтр:

```txt
service.name : "order-monitoring-api-fixed"
```

Для сравнения можно оставить тот же набор полей и показать, что ошибки стали контролируемыми HTTP-ответами.

## Сценарий 3: race condition / double charge

В Postman запустите папку:

```txt
Race Condition Demo · Concurrent Payments
```

Сценарий сам выполнит:

1. Создание заказа в broken-version.
2. Одновременную отправку нескольких `POST /orders/:id/payments`.
3. Проверку списка платежей в broken-version.
4. Создание заказа в fixed-version.
5. Такой же burst оплат.
6. Проверку списка платежей в fixed-version.

Что показывать:

- В broken-version для одного заказа обычно появляется больше одного платежа.
- В fixed-version остается один платеж, остальные запросы получают `409 order already paid`.
- В Postman Console будут статусы concurrent-запросов:

```txt
Broken concurrent payment statuses: ...
Fixed concurrent payment statuses: ...
```

## Если нужно буквально менять localhost

Вместо отдельных папок можно показывать это как переключение адреса:

```txt
broken_base_url = http://localhost:8080
fixed_base_url = http://localhost:8081
```

Сначала запускаете broken-папку, затем запускаете fixed-папку. Сценарии специально сделаны похожими, чтобы на презентации было видно: меняется только версия API, а пользовательский поток остается тем же.
