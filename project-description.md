# Order Monitoring API

## Стек

- Go
- Gin
- GORM
- PostgreSQL
- Docker Compose
- Sentry Go SDK
- Sentry Gin middleware
- Elastic Stack / Elasticsearch
- Structured logging через `log/slog`
- Postman для демонстрационных сценариев

## Краткое описание проекта

Проект демонстрирует мониторинг backend API на примере сервиса заказов. В репозитории есть две версии одного приложения:

- `broken-version` - версия с намеренно допущенными ошибками;
- `fixed-version` - версия с исправлениями, валидацией и более надежной обработкой ошибок.

Основная идея: сначала запустить broken-версию, показать ошибки в Postman, Sentry и Elastic, затем запустить fixed-версию и выполнить тот же сценарий с исправленным поведением.

## API

Эндпоинты одинаковые для обеих версий.

| Method | Endpoint | Описание |
| --- | --- | --- |
| `GET` | `/health` | Проверка доступности API |
| `GET` | `/products` | Получение списка товаров |
| `POST` | `/orders` | Создание заказа |
| `GET` | `/orders/:id` | Получение заказа по ID |
| `POST` | `/orders/discount` | Расчет скидки |
| `POST` | `/orders/:id/payments` | Захват оплаты по заказу |
| `GET` | `/orders/:id/payments` | Получение платежей по заказу |

Маршруты подключаются в `internal/router/router.go`:

```go
r.GET("/health", func(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
})

r.GET("/products", productHandler.List)
r.POST("/orders", orderHandler.Create)
r.GET("/orders/:id", orderHandler.GetByID)
r.POST("/orders/discount", orderHandler.CalculateDiscount)
r.POST("/orders/:id/payments", paymentHandler.Capture)
r.GET("/orders/:id/payments", paymentHandler.ListByOrderID)
```

## Интеграция Sentry

Sentry подключен для фиксации ошибок, panic и tracing. Если `SENTRY_DSN` не указан, интеграция отключается без падения приложения.

```go
err := sentry.Init(sentry.ClientOptions{
	Dsn:              cfg.SentryDSN,
	Environment:      cfg.SentryEnvironment,
	EnableTracing:    true,
	TracesSampleRate: cfg.SentryTracesSampleRate,
})
```

В Gin используется Sentry middleware:

```go
r.Use(sentrygin.New(sentrygin.Options{
	Repanic:         true,
	WaitForDelivery: false,
}))
```

Ошибки отправляются вручную через helper:

```go
func CaptureException(c *gin.Context, err error) {
	if err == nil {
		return
	}

	if hub := sentrygin.GetHubFromContext(c); hub != nil {
		hub.CaptureException(err)
		return
	}

	sentry.CaptureException(err)
}
```

## Интеграция Elastic Stack

Приложение пишет structured JSON-логи через `slog`. Логи идут в stdout и, если включен `ELASTIC_ENABLED=true`, дополнительно отправляются в Elasticsearch.

Основные поля HTTP-логов:

```go
fields := []any{
	"http.method", c.Request.Method,
	"url.path", path,
	"http.status_code", status,
	"event.duration_ms", latency.Milliseconds(),
	"client.address", c.ClientIP(),
	"user_agent.original", c.Request.UserAgent(),
}
```

Логгер добавляет общие поля сервиса:

```go
logger := slog.New(handler).With(
	"service.name", opts.AppName,
	"deployment.environment", opts.Environment,
)
```

Отправка в Elasticsearch выполняется в дневной индекс:

```go
index := fmt.Sprintf("%s-%s", s.indexPrefix, time.Now().UTC().Format("2006.01.02"))
res, err := s.client.Index(
	index,
	bytes.NewReader(payload),
	s.client.Index.WithContext(ctx),
)
```

В Elastic удобно сравнивать версии по полю `service.name`:

```txt
service.name : "order-monitoring-api-broken"
service.name : "order-monitoring-api-fixed"
```

## Ошибка 1: одинаковый order_number

### Broken code

Файл: `broken-version/internal/order/service.go`

```go
orderNumber := req.OrderNumber
if orderNumber == "" {
	orderNumber = "ORDER-DEMO-DUPLICATE"
}
```

Если клиент не передает `order_number`, backend всегда ставит одно и то же значение. Первый заказ создается, второй падает на unique constraint в PostgreSQL и возвращает серверную ошибку.

### Fixed code

Файл: `fixed-version/internal/order/service.go`

```go
orderNumber := strings.TrimSpace(req.OrderNumber)
if orderNumber == "" {
	orderNumber = generateOrderNumber()
}
```

```go
func generateOrderNumber() string {
	return fmt.Sprintf("ORD-%d", time.Now().UnixNano())
}
```

Дополнительно duplicate key превращается в контролируемый HTTP `409`:

```go
if err := s.orders.Create(ctx, order); err != nil {
	if isDuplicateKey(err) {
		return nil, &AppError{Status: 409, Message: "order_number already exists", Err: err}
	}
	return nil, err
}
```

Исправление: backend генерирует уникальный номер заказа, а конфликт по номеру возвращается как понятная бизнес-ошибка, а не как неожиданный `500`.

## Ошибка 2: division by zero при расчете скидки

### Broken code

Файл: `broken-version/internal/order/service.go`

```go
func (s *Service) CalculateDiscount(req DiscountRequest) DiscountResponse {
	discountAmount := req.Price / req.DiscountPercent
	finalPrice := req.Price - discountAmount

	return DiscountResponse{
		Price:           req.Price,
		DiscountPercent: req.DiscountPercent,
		DiscountAmount:  discountAmount,
		FinalPrice:      finalPrice,
	}
}
```

Если `discount_percent = 0`, приложение получает panic `integer divide by zero`. Это хорошо видно в Sentry как issue со stack trace.

### Fixed code

Файл: `fixed-version/internal/order/service.go`

```go
func (s *Service) CalculateDiscount(req DiscountRequest) (DiscountResponse, error) {
	if req.Price <= 0 {
		return DiscountResponse{}, &AppError{Status: 400, Message: "price must be greater than zero"}
	}
	if req.DiscountPercent <= 0 || req.DiscountPercent > 100 {
		return DiscountResponse{}, &AppError{Status: 400, Message: "discount_percent must be between 1 and 100"}
	}

	discountAmount := req.Price * req.DiscountPercent / 100
	finalPrice := req.Price - discountAmount
```

Исправление: добавлена валидация входных данных, формула скидки исправлена на `price * discount_percent / 100`, некорректный запрос возвращает HTTP `400`.

## Ошибка 3: искусственно медленный запрос

### Broken code

Файл: `broken-version/internal/order/service.go`

```go
if req.SimulateSlow {
	time.Sleep(3 * time.Second)
}
```

Клиент может передать `simulate_slow=true`, и API намеренно задержит обработку на 3 секунды. Это создает заметный всплеск latency в Elastic и Sentry performance/tracing.

### Fixed code

Файл: `fixed-version/internal/order/service.go`

```go
if req.SimulateSlow {
	return nil, &AppError{Status: 400, Message: "simulate_slow is disabled in fixed version"}
}
```

Исправление: опасный демонстрационный режим отключен. Запрос завершается быстро и возвращает контролируемый HTTP `400`.

## Ошибка 4: race condition при оплате

### Broken code

Файл: `broken-version/internal/payment/service.go`

```go
order, err := s.orders.FindByID(ctx, orderID)
if err != nil {
	return nil, err
}

if order.PaymentStatus == ordermodule.PaymentStatusPaid {
	return nil, ErrOrderAlreadyPaid
}

time.Sleep(250 * time.Millisecond)

payment := &Payment{
	OrderID:           order.ID,
	Amount:            order.TotalPrice,
	Status:            StatusCaptured,
	ProviderReference: generateProviderReference(),
}
```

Проблема: несколько параллельных запросов могут одновременно прочитать заказ как `pending`, пройти проверку и создать несколько платежей для одного заказа.

### Fixed code

Файл: `fixed-version/internal/payment/service.go`

```go
err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
	orderRepo := s.orders.WithDB(tx)
	paymentRepo := s.payments.WithDB(tx)

	order, err := orderRepo.FindByIDForUpdate(ctx, orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &AppError{Status: 404, Message: "order not found", Err: err}
		}
		return err
	}

	if order.PaymentStatus == ordermodule.PaymentStatusPaid {
		return &AppError{Status: 409, Message: "order already paid"}
	}
```

```go
if err := paymentRepo.Create(ctx, payment); err != nil {
	return err
}

return orderRepo.SetPaymentState(ctx, order.ID, order.PaidAmount+payment.Amount, ordermodule.PaymentStatusPaid)
```

Исправление: оплата выполняется внутри транзакции, заказ читается с блокировкой `FOR UPDATE`, повторные конкурентные запросы получают HTTP `409`, а в базе остается один платеж.

## Демонстрация через Postman

В корне проекта есть коллекция:

```txt
order-monitoring.postman_collection.json
```

Основные сценарии:

- `Presentation Runner · Broken Load` - имитация клиентов и показ ошибок broken-версии;
- `Presentation Runner · Fixed Load` - тот же сценарий на fixed-версии;
- `Race Condition Demo · Concurrent Payments` - демонстрация double charge и исправления race condition.

Подробная инструкция находится в:

```txt
postman-presentation-scenarios.md
```
