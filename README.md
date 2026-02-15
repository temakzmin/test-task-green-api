# Green API Backend

Go backend-сервис для тестового задания GREEN-API. Сервис проксирует 4 метода GREEN-API, добавляет валидацию, нормализацию `chatId`, retry/circuit breaker, JSON-логирование и graceful shutdown.

## Что реализовано

- `POST /api/v1/settings`
- `POST /api/v1/state`
- `POST /api/v1/send-message`
- `POST /api/v1/send-file-by-url`
- `GET /health`
- `GET /openapi.yaml`
- `GET /docs/index.html`

## Быстрый старт

1. Создайте локальный конфиг:

```bash
cp config/example-config.yaml config/config.yaml
```

2. Укажите порт в `config/config.yaml` (например, `5050`).

3. Запустите сервис:

```bash
make run
```

## Запуск через Docker Compose

1. Подготовьте локальный конфиг:

```bash
cp config/example-config.yaml config/config.yaml
```

2. Поднимите сервисы:

```bash
docker compose -f docker/docker-compose.yml up --build
```

3. Доступные адреса:

- frontend: `http://localhost:5000`
- backend: `http://localhost:5050`
- swagger: `http://localhost:5050/docs/index.html`

Остановить сервисы:

```bash
docker compose -f docker/docker-compose.yml down
```

## Полезные команды

```bash
make help
make test
make coverage
make check
make tidy
make hooks-install
make compose-up
make compose-down
```

## Документация API

При запущенном backend:

- OpenAPI: `http://localhost:5050/openapi.yaml`
- Swagger UI: `http://localhost:5050/docs/index.html`

Если у вас другой порт, замените `5050` на значение из `server.port`.

## Конфиг

- Пример: `config/example-config.yaml`
- Локальный: `config/config.yaml` (игнорируется git)

Ключевые параметры:

- `server.*`
- `cors.allowed_origins`
- `green_api.base_url`
- `green_api.retry.*`
- `green_api.circuit_breaker.*`
- `logging.*`

## Тесты

Запуск всех тестов:

```bash
make test
```

Запуск с покрытием:

```bash
make coverage
```

## Git Hooks (Lefthook)

В репозитории настроены хуки через `lefthook`:

- `pre-commit`: `gofumpt` + `goimports` (для staged Go-файлов), `golangci-lint run`
- `pre-push`: `make test`

Установка:

```bash
# например через go install
go install github.com/evilmartians/lefthook@latest

# установка git hooks для репозитория
make hooks-install
```

Если после `go install ...` команда `lefthook` не находится, обычно проблема в `PATH`.
Частый случай: `GOBIN` пустой, тогда бинарники ставятся в `$(go env GOPATH)/bin`.

Проверка:

```bash
go env GOBIN
go env GOPATH
```

Для `zsh` добавьте в `PATH`:

```bash
echo 'export PATH="$PATH:$(go env GOPATH)/bin"' >> ~/.zshrc
source ~/.zshrc
```

Далее повторите:

```bash
which lefthook
make hooks-install
```

## Дополнительная документация

- Деплой: `docs/deploy.md`
- Архитектура: `docs/architecture.md`
- Эксплуатация (runbook): `docs/operations.md`
- Безопасность: `docs/security.md`
- Тестирование: `docs/testing.md`
