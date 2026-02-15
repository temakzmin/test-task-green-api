# Security Notes

## 1. Secrets Handling

- Не коммитьте реальные `apiTokenInstance`.
- `config/config.yaml` должен оставаться в `.gitignore`.
- Используйте `config/example-config.yaml` только как шаблон.

## 2. Logging Policy

- Никогда не логировать полный token.
- Логировать только технические поля (status, route, latency, request_id).
- Логи должны быть в JSON формате.

## 3. CORS and Exposure

- Разрешайте только нужные `allowed_origins`.
- Не используйте `*` в production.
- Публикуйте наружу только порты, которые реально нужны.

## 4. Nginx Front Proxy

Рекомендуется:

- TLS termination на внешнем Nginx.
- Проксирование `/api/` на backend, `/` на frontend.
- Ограничения по размеру тела запроса и базовый rate limiting.

## 5. Input Validation

Backend должен валидировать все входные поля:

- `idInstance`, `apiTokenInstance` обязательны.
- `chatId` нормализуется в `@c.us`.
- `urlFile` должен быть `http/https` URL.
- `fileName` извлекается и проверяется backend.
