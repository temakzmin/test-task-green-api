# Testing Guide

## 1. Test Types

Проект использует:

- Unit tests (валидация, нормализация, retry/circuit-breaker behavior).
- Integration tests (HTTP handlers + service + greenapi client).

## 2. How to Run

```bash
make test
make coverage
```

Или напрямую:

```bash
go test ./...
go test ./... -cover
```

## 2.1 Git hooks bootstrap (Lefthook)

```bash
go install github.com/evilmartians/lefthook@latest
make hooks-install
```

Если `make hooks-install` пишет `lefthook is not installed`, проверьте `PATH`.
Когда `GOBIN` пустой, бинарник ставится в `$(go env GOPATH)/bin`:

```bash
echo 'export PATH="$PATH:$(go env GOPATH)/bin"' >> ~/.zshrc
source ~/.zshrc
```

## 2.2 Frontend Local Run Without Reverse Proxy

По умолчанию в фронтенде используется `API_BASE = "/api/v1"`.
Это работает в production, когда серверный reverse proxy направляет `/api/*` в backend.

Если вы запускаете frontend локально без reverse proxy (например, `http://localhost:5000`), то запросы `POST /api/v1/*` уйдут в статический сервер frontend и вернут `501`.

Для локального тестирования в таком режиме временно измените `API_BASE` в `frontend/app.js` на:

```js
const API_BASE = "http://localhost:5050/api/v1";
```

## 3. Coverage Goal

Целевой диапазон: 70-90% по backend-коду.

Примечание: важнее покрыть критические ветки (валидация, ошибки upstream, retry/circuit breaker), чем добиваться максимума любой ценой.

## 4. Naming and Layout

- Файлы тестов: `*_test.go`.
- Имена тестов: `Test<Subject>_<Scenario>`.
- Тестовые кейсы группируются по package-слоям (`config`, `service`, `greenapi`, `http/router`).

## 5. Pre-merge Checklist

Перед merge:

- `make test` проходит.
- `make coverage` показывает адекватное покрытие ключевых сценариев.
- Новая бизнес-логика сопровождается unit и/или integration тестом.
