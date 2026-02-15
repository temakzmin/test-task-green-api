# Deploy Guide

## 1. Prerequisites

На сервере должны быть установлены:
- Docker Engine
- Docker Compose plugin (`docker compose`)
- Nginx (вне Docker)
- Домен, направленный на сервер

## 2. Prepare Project

Склонируйте репозиторий и перейдите в директорию проекта:

```bash
git clone <your-repo-url> green-api
cd green-api
```

Создайте рабочий конфиг backend:

```bash
cp config/example-config.yaml config/config.yaml
```

Проверьте `config/config.yaml`:
- `server.port: 5050`
- `cors.allowed_origins`: ваш frontend origin
- `green_api.base_url`: актуальный URL GREEN-API

## 3. Start Services (Standard)

Запустите backend и frontend через Docker Compose:

```bash
docker compose -f docker/docker-compose.yml up --build -d
```

Проверьте, что контейнеры поднялись:

```bash
docker compose -f docker/docker-compose.yml ps
```

## 4. Start Services (Weak VPS: Prebuilt Backend Binary)

Если VPS слабый и не должен собирать backend, соберите Linux-бинарник локально и загрузите его на сервер.

На локальной машине (или CI):

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/server-linux-amd64 ./cmd/server
```

Передайте бинарник на сервер (пример):

```bash
scp build/server-linux-amd64 user@your-server:/opt/green-api/build/server-linux-amd64
```

На сервере:

```bash
cd /opt/green-api
make compose-prebuilt-up
```

Пояснение:
- `make compose-prebuilt-up` проверяет наличие `build/server-linux-amd64` и запускает `docker compose` с `docker/docker-compose.prebuilt.yml`.
- В prebuilt-сценарии backend-контейнер использует `docker/Dockerfile.backend.prebuilt` и только копирует готовый бинарник в образ, без сборки Go внутри Docker на сервере.

## 5. Configure Nginx

Пример `server` блока Nginx (один домен для фронта и API):

```nginx
server {
    listen 80;
    server_name your-domain.com;

    location /api/ {
        proxy_pass http://127.0.0.1:5050;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /docs {
        proxy_pass http://127.0.0.1:5050;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /openapi.yaml {
        proxy_pass http://127.0.0.1:5050;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location / {
        proxy_pass http://127.0.0.1:5000;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

Проверьте и примените конфиг:

```bash
sudo nginx -t
sudo systemctl reload nginx
```

## 6. Smoke Checks

Проверьте доступность:
- `http://your-domain.com/`
- `http://your-domain.com/api/v1/settings` (через POST)
- `http://your-domain.com/docs`
- `http://your-domain.com/openapi.yaml`

Логи контейнеров:

```bash
docker compose -f docker/docker-compose.yml logs -f
```

## 7. Update Procedure

Обновление приложения:

```bash
git pull
docker compose -f docker/docker-compose.yml up --build -d
```

Обновление для сценария слабого VPS (prebuilt):
1. Собрать новый Linux-бинарник на локальной машине/CI.
2. Передать `build/server-linux-amd64` на сервер.
3. На сервере выполнить `make compose-prebuilt-up`.

Rollback (пример):
- переключиться на нужный commit/tag
- повторить `docker compose ... up --build -d`
