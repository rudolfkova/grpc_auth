WIP
Нужно доделать:
- docker
- тесты для chat-service и gateway
- конфиги в .env
- дополнительное логирование
- метрики
- сервис с профилями
# Messenger

Пет-проект: мессенджер на Go с микросервисной архитектурой, gRPC и real-time доставкой сообщений через WebSocket.

## Сервисы

| Сервис | Описание | Порт |
|---|---|---|
| **auth-service** | Регистрация, логин, JWT + refresh токены, сессии | `:50051` |
| **chat-service** | Чаты, сообщения, real-time push через gRPC stream | `:50052` |
| **gateway** | REST HTTP + WebSocket фасад над gRPC сервисами | `:8080` |
Порты можно менять в конфигах. В будущем планируется перенос из .toml в .env чтобы было проще деплоить.

## Стек

- **Go** - gRPC, net/http, database/sql
- **PostgreSQL** - основное хранилище (оба сервиса, отдельные БД)
- **Redis** - кэш сессий в auth-service
- **JWT** - access + refresh токены (HS256)
- **WebSocket** - real-time доставка сообщений через gorilla/websocket
- **protobuf / gRPC** - межсервисное взаимодействие

## Архитектура

```
Client (browser) -> WebSocket -> Gateway -> chat-service Subscribe (gRPC stream)
|
REST HTTP
|
Gateway
|
auth-service (gRPC)
chat-service (gRPC)
```

**auth-service** выдаёт JWT. **chat-service** валидирует его локально (подпись + exp) и проверяет активность сессии через `ValidateSession` в auth-service. `user_id` из верифицированного токена передаётся через контекст - бизнес-логика не доверяет данным из запроса.

Real-time: при отправке сообщения chat-service пушит его через in-memory Hub всем подписчикам чата. Gateway держит WebSocket соединения клиентов и транслирует сообщения из gRPC stream.

## Запуск

**Зависимости:** Go 1.22+, PostgreSQL, Redis

```bash
# Миграции
make migrate-auth-up DB_DSN="postgres://..."
make migrate-chat-up DB_DSN="postgres://..."

# Сборка
make build-auth && make build-chat && make build-gateway

# Запуск (три терминала)
make start-auth
make start-chat
make start-gateway
```

Конфигурация через `config.toml`, `config-chat.toml`, `config-gateway.toml`.

Открыть `messenger-ui.html` в браузере - готовый интерфейс для тестирования.
