# Architecture Overview

## 1. High-level Flow

Запрос проходит цепочку:

`Frontend -> Backend (Gin) -> Service -> GreenAPI Client -> GREEN-API`

Ответ возвращается обратно по той же цепочке. Backend не хранит состояние сессий и работает как stateless proxy/adapter.

## 2. Backend Layers

- `internal/http/handler`: HTTP endpoints `/api/v1/*`, bind/response, mapping ошибок.
- `internal/service`: валидация и бизнес-правила (`chatId` normalization, `fileName` extraction).
- `internal/greenapi`: интеграция с внешним GREEN-API, retry, circuit breaker.
- `internal/middleware`: `request_id`, request logging.
- `internal/config`: загрузка и валидация YAML-конфига.
- `internal/logging`: инициализация JSON logger.

## 3. Reliability Model

- Retry только на network/timeout/HTTP 5xx.
- На HTTP 4xx retry не выполняется.
- Circuit breaker (`closed/open/half-open`) защищает backend от деградации upstream.
- Таймауты backend и graceful shutdown конфигурируемы.

## 4. API Contract

Публичные backend endpoints:

- `POST /api/v1/settings`
- `POST /api/v1/state`
- `POST /api/v1/send-message`
- `POST /api/v1/send-file-by-url`

Документация контракта:

- `GET /openapi.yaml`
- `GET /docs` (`/docs/index.html`)

## 5. Observability

- JSON logs (zap).
- В каждом запросе `request_id` (header `X-Request-Id`).
- Логи содержат route/method/status/latency/request_id.
