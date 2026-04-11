# Лобби в Postman: чат и мир

Порты по умолчанию: **gateway** `http://localhost:8080`, **world-service (gRPC)** `localhost:50054`. Стек должен быть запущен (`docker compose up` или локально).

---

## 1. Токен (для чата через gateway)

Чат в **chat-service** требует JWT. Удобнее идти через **gateway**:

1. **POST** `http://localhost:8080/auth/register`  
   Body (JSON): `{ "email": "lobby@test.local", "password": "secret" }`

2. **POST** `http://localhost:8080/auth/login`  
   Body: `{ "email": "lobby@test.local", "password": "secret", "app_id": 1 }`  
   Из ответа скопируй **`access_token`**.

3. Дальше к чату: заголовок **`Authorization: Bearer <access_token>`** (одна строка, с пробелом после `Bearer`).

---

## 2. Создать чат (лобби-канал)

**HTTP (Postman), не gRPC.**

| Поле | Значение |
|------|----------|
| Метод | **POST** |
| URL | `http://localhost:8080/chat/create` |
| Headers | `Content-Type: application/json`, `Authorization: Bearer <access_token>` |
| Body | `{"name": "Моё лобби"}` — `name` можно опустить или оставить пустым `""` |

Ответ: **`chat_id`**, **`created_at`**. Сохрани `chat_id`.

### Участники лобби

**POST** `http://localhost:8080/chat/members`  
Те же заголовки + JSON:

```json
{ "chat_id": 123, "user_id": 1 }
```

Повтори для каждого `user_id`, кого нужно добавить в чат (id из auth после регистрации / из своей базы).

---

## 3. Создать мир (world-service, gRPC)

Мир **не** заведён через gateway — только **gRPC** к **world-service**.

1. В Postman: **New → gRPC**, URL `localhost:50054`, без TLS (plaintext).
2. Импорт proto: `world-service/proto/world/v1/world.proto` (или server reflection, если включён).
3. Сервис **`world.v1.WorldService`**, метод **`CreateWorld`**.
4. Metadata: если в `config-world.toml` задан **`service_token`**, добавь **`x-service-token: <токен>`** (или `Authorization: Bearer <токен>`). В docker-конфиге по умолчанию токен пустой — metadata не нужны.

### Тело `CreateWorldRequest`

Порядок полей как в proto: **`id`**, **`name`**, **`description`**, **`snapshot`**, **`schema_version`**.

Пример (JSON; **`snapshot`** — base64, пустой мир — `""`; те же ключи можно писать как в proto: `schema_version`).

```json
{
  "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "name": "lobby-world",
  "description": "Лобби",
  "snapshot": "",
  "schema_version": 1
}
```

- **`id`**: свой UUID мира (один раз сгенерируй, потом его же подставь в **`WORLD_ID` / `world_id`** в `game-service`).
- **`schemaVersion`**: начни с **`1`**.

Ответ: объект **`world`** с тем же **`id`**, **`version`**, и т.д.

Проверка: **gRPC** `WorldService` → **`GetWorld`**, body `{ "id": "<тот же uuid>" }`.

---

## 4. Связка лобби

| Сущность | Зачем |
|----------|--------|
| **`chat_id`** | Клиенты / gateway: чат лобби, участники, сообщения. |
| **`world` `id`** | **game-service**: `world_id` в `config-game.toml` или `WORLD_ID` в окружении — тот же UUID, что в `CreateWorld`. |

---

## 5. Замечание про Postman и JSON (gRPC)

Если в gRPC-запросах поля «не доезжают», выставь ключи в JSON **в том же порядке**, что поля в `message` в `.proto` (как в `character-service/POSTMAN_TESTS.md`).

---

## 6. Альтернатива чату: DM без имени

**POST** `http://localhost:8080/chat/get-or-create`  
Headers: как у создания чата.  
Body:

```json
{ "initiator_id": 1, "recipient_id": 2 }
```

Ответ: **`chat_id`**, **`created`** (создан новый или взяли существующий).
