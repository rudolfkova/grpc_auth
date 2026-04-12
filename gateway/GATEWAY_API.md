# Gateway HTTP/WebSocket API

Документ описывает **публичный интерфейс gateway**: как вызывать ручки с клиента или из скриптов. Базовый URL зависит от деплоя (например `http://localhost:8080` — смотри `config-gateway.toml`, поле `bind_addr`).

## Общие правила

| Правило | Значение |
|--------|-----------|
| Формат тела запросов | JSON, заголовок `Content-Type: application/json` |
| Формат успешных ответов | JSON |
| Ошибки | Обычно JSON `{"error": "<текст>"}`; для **`GET /api/me/characters`** — стабильное тело `{"code":"<машинное имя>","message":"<текст>"}` (см. раздел ниже). HTTP-статус по смыслу (400, 401, 404, 409, 429, 502, 503, 504 и т.д.). |
| CORS | `Access-Control-Allow-Origin: *`, методы `GET, POST, PUT, DELETE, OPTIONS`, заголовки `Content-Type, Authorization`. Для preflight шлите `OPTIONS` — ответ `204 No Content`. |
| Даты в JSON | Поля времени сериализуются как строки **RFC3339** (например `2026-04-09T15:04:05+03:00`). |

### Авторизация для чата

Эндпоинты под префиксом `/chat/...` (кроме случаев, где указано иначе) ожидают заголовок:

```http
Authorization: Bearer <access_token>
```

Значение — **access token** из ответа `POST /auth/login` или `POST /auth/refresh`. Сам заголовок целиком прокидывается в gRPC как метаданные `authorization` (с префиксом `Bearer ` внутри строки, если клиент так передал).

Если токен невалиден или отсутствует, backend может вернуть **401** с телом `{"error":"..."}`.

### Эндпоинт без токена в заголовке

- `GET /auth/is-admin?user_id=...` — проверка флага админа по **числовому** `user_id` в query (отдельно от JWT в этом gateway).

---

## Персонажи для игры (`GET /api/me/characters`)

Игровой клиент после логина (есть **access JWT**) может **без gRPC и без `service_token`** получить список персонажей текущего пользователя перед WebSocket **`/ws/game`**.

### `GET /api/me/characters`

**Идемпотентно**, только **GET**.

**Заголовок (обязательно):**

```http
Authorization: Bearer <access_token>
```

Тот же access token, что из **`POST /auth/login`** / **`POST /auth/refresh`**. Секрет подписи JWT на gateway должен совпадать с auth-service (**`jwt_secret`** в `config-gateway.toml`).

**Query (опционально):**

| Параметр | По умолчанию | Описание |
|----------|---------------|----------|
| `limit`  | `50`          | Максимум **`100`** (больше — режется до 100). Меньше `1` игнорируется, остаётся 50. |
| `offset` | `0`           | Смещение для пагинации (неотрицательное целое). |

**Поведение:** gateway валидирует JWT → из claims читает **`user_id`** → вызывает character-service **`ListCharacters(user_id, limit, offset)`** с внутреннего процесса (при настройке — metadata **`x-service-token`** из **`character_service_token`**).

**Успешный ответ `200`:**

```json
{
  "characters": [
    {
      "id": "<uuid>",
      "display_name": "...",
      "description": "...",
      "schema_version": 2,
      "version": 7
    }
  ]
}
```

Поле **`data`** (blob персонажа) **не отдаётся** в листинге (меньше трафика). Детали blob — в **`character-service/CHARACTER_EDITOR_CLIENT.md`**; при необходимости позже можно добавить **`GET /api/me/characters/:id`**.

**Ошибки (тело `{"code","message"}`):**

| HTTP | `code` | Когда |
|------|--------|--------|
| `401` | `unauthorized` | Нет/битый Bearer, истёкший или невалидный JWT, в claims нет `user_id`. |
| `405` | `method_not_allowed` | Не GET. |
| `503` | `character_service_unconfigured` | В конфиге gateway пустой **`character_service_addr`** (эндпоинт отключён). |
| `503` | `character_service_unavailable` | gRPC **`Unavailable`** от character-service (сеть, сервис не поднят). |
| `504` | `character_service_timeout` | **`DeadlineExceeded`**. |
| `400` | `character_service_invalid_argument` | Некорректные аргументы со стороны character-service. |
| `502` | `character_service_error` / `character_upstream_error` | Прочие ошибки gRPC или не-gRPC при вызове upstream. |

**Связка с игрой:** после выбора UUID из массива откройте **`GET /ws/game?token=<access_token>&character_id=<uuid>`** (см. **`game-service/README.md`**). Gateway пробрасывает **`character_id`** в game-service.

**Конфиг gateway:** `character_service_addr`, `character_service_token` (опционально, если на character-service включён `service_token`).

---

## Здоровье

### `GET /health`

Без тела запроса.

**Ответ `200`:**

```json
{ "status": "ok" }
```

---

## Аутентификация (`/auth/...`)

### `POST /auth/register`

Создание пользователя.

**Тело:**

```json
{
  "email": "user@example.com",
  "password": "secret"
}
```

Поля `email` и `password` обязательны.

**Ответ `201`:**

```json
{ "user_id": 42 }
```

**Возможные ошибки:** 400 (валидация), 409 (email уже занят — приходит как gRPC `AlreadyExists`).

---

### `POST /auth/login`

**Тело:**

```json
{
  "email": "user@example.com",
  "password": "secret",
  "app_id": 1
}
```

`email` и `password` обязательны. `app_id` опционален; если **0** или не передан, gateway подставляет **1**.

**Ответ `200`:**

```json
{
  "access_token": "<jwt>",
  "refresh_token": "<string>",
  "access_expires_at": "2026-04-09T16:00:00+03:00",
  "refresh_expires_at": "2026-04-16T15:00:00+03:00"
}
```

Поля с датами могут быть `null`, если сервер не заполнил время.

**Ошибки:** 401 при неверных учётных данных.

---

### `POST /auth/logout`

Инвалидирует сессию по refresh-токену.

**Тело:**

```json
{ "refresh_token": "<refresh_token из login/refresh>" }
```

**Ответ `200`:**

```json
{ "success": true }
```

(`success` приходит из gRPC; трактуйте как булево.)

---

### `POST /auth/refresh`

**Тело:**

```json
{ "refresh_token": "<refresh_token>" }
```

**Ответ `200`:** тот же JSON, что у **login** (`access_token`, `refresh_token`, даты).

---

### `GET /auth/is-admin`

**Query (обязательно):** `user_id=<int64>`

Пример: `GET /auth/is-admin?user_id=42`

**Ответ `200`:**

```json
{ "is_admin": false }
```

Токен в заголовке для этой ручки **не требуется** (в текущей реализации gateway).

---

## Чат (`/chat/...`)

Все перечисленные ниже методы, кроме явных исключений, требуют заголовок **`Authorization: Bearer <access_token>`**.

Тела у **`DELETE`** передаются как **JSON в body** (не все HTTP-клиенты делают это по умолчанию — включите body вручную).

---

### `POST /chat/create`

Создать чат.

**Тело:**

```json
{ "name": "Название чата" }
```

`name` можно опустить или оставить пустым — тогда создаётся заготовка под DM без участников (см. комментарии в коде gateway).

**Ответ `200`:**

```json
{
  "chat_id": 1,
  "created_at": "2026-04-09T15:00:00+03:00"
}
```

---

### `DELETE /chat`

Удалить чат.

**Тело:**

```json
{ "chat_id": 1 }
```

**Ответ `200`:** `{ "ok": true }`

---

### `POST /chat/members`

Добавить участника.

**Тело:**

```json
{
  "chat_id": 1,
  "user_id": 2
}
```

**Ответ `200`:** `{ "ok": true }`

---

### `DELETE /chat/members`

Удалить участника.

**Тело:**

```json
{
  "chat_id": 1,
  "user_id": 2
}
```

**Ответ `200`:** `{ "ok": true }`

---

### `POST /chat/get-or-create`

Найти или создать диалог между двумя пользователями.

**Тело:**

```json
{
  "initiator_id": 1,
  "recipient_id": 2
}
```

Оба поля обязательны и не должны быть 0.

**Ответ `200`:**

```json
{
  "chat_id": 10,
  "created": true,
  "created_at": "2026-04-09T15:00:00+03:00"
}
```

`created` — был ли чат создан в этом запросе (семантика chat-service).

---

### `GET /chat/messages`

История сообщений с пагинацией.

**Query:**

| Параметр   | Обязательно | Описание |
|------------|-------------|----------|
| `chat_id`  | да          | ID чата |
| `limit`    | нет         | По умолчанию **50** |
| `cursor`   | нет         | Курсор следующей страницы (строка из прошлого ответа `next_cursor`) |

Пример: `GET /chat/messages?chat_id=1&limit=20`

**Ответ `200`:**

```json
{
  "messages": [
    {
      "id": 100,
      "chat_id": 1,
      "sender_id": 2,
      "text": "hello",
      "created_at": "2026-04-09T15:00:00+03:00"
    }
  ],
  "next_cursor": ""
}
```

`next_cursor` — пустая строка или токен для следующего запроса.

---

### `GET /chat/chats`

Список чатов пользователя.

**Query:**

| Параметр  | Обязательно | Описание |
|-----------|-------------|----------|
| `user_id` | да          | ID пользователя |
| `limit`   | нет         | По умолчанию **50** |
| `offset`  | нет         | Смещение, по умолчанию **0** |

Пример: `GET /chat/chats?user_id=1&limit=10&offset=0`

**Ответ `200`:**

```json
{
  "chats": [
    {
      "chat_id": 1,
      "name": "DM",
      "companion_id": 2,
      "last_message": "hi",
      "unread_count": 0,
      "last_message_at": "2026-04-09T15:00:00+03:00"
    }
  ]
}
```

---

### `POST /chat/send`

Отправить сообщение.

**Тело:**

```json
{
  "chat_id": 1,
  "sender_id": 2,
  "text": "Текст сообщения"
}
```

Все три поля обязательны; `text` не должен быть пустым.

**Ответ `200`:**

```json
{
  "message_id": 200,
  "created_at": "2026-04-09T15:00:00+03:00"
}
```

---

## WebSocket: чат

### `GET /ws/subscribe`

Подписка на **поток новых сообщений** через chat-service (gRPC стрим → JSON в WebSocket).

**Подключение:** обычный WebSocket upgrade на этот URL.

**Query (важно):**

| Параметр | Описание |
|----------|----------|
| `token`  | **Access JWT** (без префикса `Bearer` в query). Gateway сам формирует для gRPC строку `Bearer <token>`. |

Пример URL: `ws://<host>/ws/subscribe?token=<access_token>`

**Исходящие сообщения (сервер → клиент):** один JSON-объект на сообщение:

```json
{
  "id": 100,
  "chat_id": 1,
  "sender_id": 2,
  "text": "hello",
  "created_at": "2026-04-09T15:00:00+03:00"
}
```

Клиент **ничего не обязан** слать по сокету после подключения (односторонний пуш из стрима).

---

## WebSocket: игра

### `GET /ws/game`

Прокси до **game-service**: после установки соединения формат сообщений такой же, как у игрового сервиса (конверты `service` / `type` / `payload`).

**Query:**

| Параметр        | Описание |
|-----------------|----------|
| `token`         | Access JWT (как в игровом WS). |
| `character_id`  | UUID персонажа (обязателен, если на game-service настроен character-service). Пробрасывается в game-service. |
| `session_id`    | Опционально; пробрасывается в backend как query-параметр. |

Пример: `ws://<host>/ws/game?token=<access_token>&character_id=<uuid>`

Если game-service недоступен, клиент может получить одно JSON-сообщение и закрытие:

```json
{
  "service": "game",
  "type": "error",
  "payload": { "message": "game service unavailable" }
}
```

Подробный протокол игры (типы `move`, `hit`, `spawn_tile`, `clear_tile`, `state`, `reject`, логирование) — в репозитории: **`game-service/README.md`** и **`pkg/gamekit/README.md`**.

---

## Карта HTTP-методов (кратко)

| Метод | Путь | Назначение |
|-------|------|------------|
| GET | `/health` | Liveness |
| POST | `/auth/register` | Регистрация |
| POST | `/auth/login` | Логин |
| POST | `/auth/logout` | Выход (refresh) |
| POST | `/auth/refresh` | Новые токены |
| GET | `/auth/is-admin` | Админ по `user_id` |
| GET | `/api/me/characters` | Список персонажей текущего пользователя (Bearer JWT → character-service) |
| POST | `/chat/create` | Создать чат |
| DELETE | `/chat` | Удалить чат (JSON body) |
| POST | `/chat/members` | Добавить участника |
| DELETE | `/chat/members` | Удалить участника (JSON body) |
| POST | `/chat/get-or-create` | DM get-or-create |
| GET | `/chat/messages` | История |
| GET | `/chat/chats` | Список чатов |
| POST | `/chat/send` | Отправить сообщение |
| GET | `/ws/subscribe` | WS чат |
| GET | `/ws/game` | WS игра (прокси) |

---

## Подсказки для клиентов

1. После **login** сохраняйте `access_token` и подставляйте в `Authorization: Bearer ...` для REST чата.
2. Для **WebSocket чата** используйте тот же access token **только** в query `token=...` (не дублируйте `Bearer` в query — gateway добавит сам для gRPC).
3. Для **WebSocket игры** передавайте `token` так же, как ожидает game-service (сырой JWT в query); при включённом character-service на game — добавьте **`character_id`** (сначала **`GET /api/me/characters`**).
4. При **429** / **503** имеет смысл повторить запрос с backoff.
5. Ошибки **`GET /api/me/characters`** логируйте по полю **`code`** в JSON.
