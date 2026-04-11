# Ручные проверки character-service в Postman (gRPC)

Сервис: **`character.v1.CharacterService`**, адрес **`localhost:50055`** (или хост:порт из docker-compose).

## Postman: порядок ключей в JSON

В ряде версий Postman gRPC тело JSON сопоставляется с полями protobuf **в том числе по порядку ключей** в объекте (как порядок полей в `message` в `.proto`). Если ключи переставить местами, значения могут попасть не в те поля (например `user_id` остаётся 0).

**Правило:** в каждом примере ниже ключи идут **в порядке объявления полей** в `character.proto` (поле 1, затем 2, …). Для целых полей используй **число** (`9001`), не строку (`"9001"`), если только Postman у тебя явно не требует иначе.

Имена ключей можно брать **как в proto** (`user_id`, `display_name`) или в стиле **proto3 JSON** (`userId`, `displayName`) — главное, чтобы порядок совпадал с порядком полей в соответствующем `message`.

---

## Подготовка

1. **New → gRPC** (или аналог в твоей версии Postman).
2. **Server URL:** `localhost:50055` (без TLS, если так настроено).
3. **Схема:** импорт  
   `character-service/proto/character/v1/character.proto`  
   либо **Server reflection**.
4. **Metadata:** если в `config-character.toml` задан `service_token` — `x-service-token` или `Authorization: Bearer …`.

## Константы

| Имя | Пример | Назначение |
|-----|--------|------------|
| `USER_ID` | `9001` | владелец (как user из auth) |
| `CHAR_ID` | `aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa` | UUID для тестов «ещё нет в БД» |

После **CreateCharacter** сохрани **`character.id`** → **`SAVED_ID`**.

---

## Тест 1 — ListCharacters

**Порядок полей в proto:** `user_id`, `limit`, `offset`.

```json
{
  "user_id": 9001,
  "limit": 50,
  "offset": 0
}
```

Эквивалент по смыслу (camelCase, тот же порядок):

```json
{
  "userId": 9001,
  "limit": 50,
  "offset": 0
}
```

**Ожидание:** OK, массив **`characters`** (может быть пустым).

---

## Тест 2 — ResolvePlayCharacter (нет строки в БД)

**Порядок:** `user_id`, `id`.

```json
{
  "user_id": 9001,
  "id": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
}
```

**Ожидание:** OK, **`persisted`: false**, в **`character`** те же идентификаторы, без данных из БД.

---

## Тест 3 — GetCharacter → NotFound

Тело как в тесте 2.

**Ожидание:** **NotFound**.

---

## Тест 4 — CreateCharacter (пустой `id` → UUID с сервера)

**Порядок в proto:** `user_id`, `id`, `display_name`, `description`, `data`, `schema_version`.

```json
{
  "user_id": 9001,
  "id": "",
  "display_name": "PostmanHero",
  "description": "manual test",
  "data": "UE9TVA==",
  "schema_version": 1
}
```

**Ожидание:** OK, непустой **`character.id`** → **`SAVED_ID`**.

---

## Тест 5 — ListCharacters после создания

Как тест 1.

**Ожидание:** в списке есть персонаж с **`id`** = **`SAVED_ID`**.

---

## Тест 6 — GetCharacter

**Порядок:** `user_id`, `id`.

```json
{
  "user_id": 9001,
  "id": "ВСТАВЬ_SAVED_ID"
}
```

**Ожидание:** OK, **`data`** соответствует созданию.

---

## Тест 7 — GetCharacterByDisplayName

**Порядок:** `user_id`, `display_name`.

```json
{
  "user_id": 9001,
  "display_name": "PostmanHero"
}
```

**Ожидание:** OK, **`character.id`** = **`SAVED_ID`**.

---

## Тест 8 — ReplaceCharacterData

**Порядок:** `user_id`, `id`, `data`, `schema_version`, `expected_version`.

```json
{
  "user_id": 9001,
  "id": "ВСТАВЬ_SAVED_ID",
  "data": "TkVX",
  "schema_version": 1,
  "expected_version": 0
}
```

**Ожидание:** OK, **`version`** вырос.

---

## Тест 9 — ResolvePlayCharacter после сохранения

Как тест 2, но **`id`**: **`SAVED_ID`**.

**Ожидание:** **`persisted`: true**.

---

## Тест 10 — DeleteCharacter

**Порядок:** `user_id`, `id`.

```json
{
  "user_id": 9001,
  "id": "ВСТАВЬ_SAVED_ID"
}
```

**Ожидание:** OK.

---

## Тест 11 — GetCharacter → NotFound

Как тест 6.

**Ожидание:** **NotFound**.

---

## Тест 12 — изоляция по владельцу

**GetCharacter:** тот же **`id`**, **`user_id`** другой (например `9002`).

**Ожидание:** **NotFound**.

---

## Справка: proto3 JSON (camelCase)

| В proto | Частый JSON-ключ |
|---------|------------------|
| `user_id` | `userId` |
| `display_name` | `displayName` |
| `schema_version` | `schemaVersion` |
| `expected_version` | `expectedVersion` |

**`bytes`** — строка **base64**.

После смены proto **переимпортируй** файл или обнови **reflection**.
