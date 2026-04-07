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

### Через Docker (одной командой)

```bash
make docker-up
```

По умолчанию docker-команды в `Makefile` запускаются через `sudo docker`, чтобы избежать ошибок доступа к daemon.
Если у вас уже настроена группа `docker`, можно запускать без `sudo`:

```bash
make DOCKER=docker docker-up
```

После старта:
- Gateway: `http://localhost:8080`
- Auth gRPC: `localhost:50051`
- Chat gRPC: `localhost:50052`

`postgres` и `redis` в docker-compose доступны только внутри docker-сети (без проброса портов на хост), чтобы не конфликтовать с локальными инстансами.

Остановка:

```bash
make docker-down
```

Полная очистка docker-стека (контейнеры + тома):

```bash
make docker-reset
```

Логи:

```bash
make docker-logs
```

Docker-конфиги лежат в `deploy/docker`.

### Локально (без Docker)

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

## API Contract

Этого раздела достаточно, чтобы использовать сервис как "чёрный ящик" и не смотреть внутрь кода.

### База

- Base URL: `http://localhost:8080`
- Формат тела: `application/json`
- Для защищённых chat-методов передавать заголовок:
  - `Authorization: Bearer <access_token>`

### Health

- `GET /health`
- Ответ `200`:

```json
{"status":"ok"}
```

---

## Auth API

### Register

- `POST /auth/register`
- Body:

```json
{
  "email": "player@example.com",
  "password": "strong_password"
}
```

- Успех: `201`

```json
{
  "user_id": 123
}
```

### Login

- `POST /auth/login`
- Body (`app_id` опционален, по умолчанию `1`):

```json
{
  "email": "player@example.com",
  "password": "strong_password",
  "app_id": 1
}
```

- Успех: `200`

```json
{
  "access_token": "<jwt>",
  "refresh_token": "<jwt>",
  "access_expires_at": "2026-04-07T21:00:00Z",
  "refresh_expires_at": "2026-05-07T21:00:00Z"
}
```

### Refresh token pair

- `POST /auth/refresh`
- Body:

```json
{
  "refresh_token": "<jwt>"
}
```

- Успех: `200` (возвращает новую пару access/refresh как в `/auth/login`)

### Logout

- `POST /auth/logout`
- Body:

```json
{
  "refresh_token": "<jwt>"
}
```

- Успех: `200`

```json
{
  "success": true
}
```

### Is Admin

- `GET /auth/is-admin?user_id=<id>`
- Успех: `200`

```json
{
  "is_admin": false
}
```

---

## Chat API

Все методы ниже ожидают `Authorization: Bearer <access_token>`.

### Get or Create DM chat

- `POST /chat/get-or-create`
- Body:

```json
{
  "initiator_id": 1,
  "recipient_id": 2
}
```

- Успех: `200`

```json
{
  "chat_id": 10,
  "created": true,
  "created_at": "2026-04-07T21:10:00Z"
}
```

### Send message

- `POST /chat/send`
- Body:

```json
{
  "chat_id": 10,
  "sender_id": 1,
  "text": "hello from game client"
}
```

- Успех: `200`

```json
{
  "message_id": 501,
  "created_at": "2026-04-07T21:11:00Z"
}
```

### Get messages (history, cursor pagination)

- `GET /chat/messages?chat_id=<id>&limit=50&cursor=<cursor>`
- `limit` опционален (по умолчанию `50`)
- `cursor` опционален
- Успех: `200`

```json
{
  "messages": [
    {
      "id": 501,
      "chat_id": 10,
      "sender_id": 1,
      "text": "hello",
      "created_at": "2026-04-07T21:11:00Z"
    }
  ],
  "next_cursor": ""
}
```

### Get user chats (offset pagination)

- `GET /chat/chats?user_id=<id>&limit=50&offset=0`
- `limit` опционален (по умолчанию `50`)
- `offset` опционален
- Успех: `200`

```json
{
  "chats": [
    {
      "chat_id": 10,
      "companion_id": 2,
      "last_message": "hello",
      "unread_count": 0,
      "last_message_at": "2026-04-07T21:11:00Z"
    }
  ]
}
```

---

## WebSocket (real-time)

- Endpoint: `GET /ws/subscribe?token=<access_token>`
- Пример URL: `ws://localhost:8080/ws/subscribe?token=<access_token>`
- Сервер пушит JSON-события новых сообщений:

```json
{
  "id": 501,
  "chat_id": 10,
  "sender_id": 1,
  "text": "hello",
  "created_at": "2026-04-07T21:11:00Z"
}
```

---

## Ошибки (единый формат)

При ошибке gateway отвечает JSON:

```json
{
  "error": "message"
}
```

Типичные HTTP-коды:

- `400` invalid request / missing params
- `401` unauthenticated (токен отсутствует/просрочен/невалиден)
- `403` permission denied
- `404` not found
- `409` already exists
- `429` rate/resource exhausted
- `503` upstream service unavailable
- `500` internal error

---

## Рекомендуемый flow

1. `POST /auth/register` (один раз) или сразу `POST /auth/login`
2. Сохранить `access_token` и `refresh_token`
3. Открыть WS: `/ws/subscribe?token=<access_token>`
4. Создавать/получать чат: `POST /chat/get-or-create`
5. Отправлять сообщения: `POST /chat/send`
6. Подтягивать историю: `GET /chat/messages`
7. При `401` делать `POST /auth/refresh` и повторять запрос
8. При выходе пользователя: `POST /auth/logout`
