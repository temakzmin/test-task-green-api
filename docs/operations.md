# Operations Runbook

## 1. Start / Stop

Локально через Docker Compose:

```bash
docker compose -f docker/docker-compose.yml up --build -d
docker compose -f docker/docker-compose.yml down
```

## 2. Health and Docs Checks

- Health: `GET /health`
- OpenAPI: `GET /openapi.yaml`
- Swagger UI: `GET /docs`

Примеры:

```bash
curl -s http://localhost:5050/health
curl -I http://localhost:5050/docs
```

## 3. Logs and Diagnostics

Логи контейнеров:

```bash
docker compose -f docker/docker-compose.yml logs -f backend
docker compose -f docker/docker-compose.yml logs -f frontend
```

Что проверять при ошибках:

- `400`: ошибки валидации payload.
- `502`: проблемы связи с GREEN-API (network/upstream error).
- `503`: circuit breaker в состоянии `open`.
- `504`: таймаут вызова upstream.

## 4. Common Incidents

### 4.1 CORS errors in browser

Проверьте `cors.allowed_origins` в `config/config.yaml` и перезапустите backend.

### 4.2 GREEN-API unavailable

Проверьте `green_api.base_url`, сеть и токены инстанса.

### 4.3 Circuit breaker keeps open

- проверьте стабильность upstream;
- временно увеличьте `open_timeout_seconds` и/или пороги;
- уменьшите нагрузку до восстановления upstream.

## 5. Update Procedure

```bash
git pull
docker compose -f docker/docker-compose.yml up --build -d
```

При необходимости rollback:

```bash
git checkout <commit-or-tag>
docker compose -f docker/docker-compose.yml up --build -d
```
