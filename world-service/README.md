# world-service

Внутренний gRPC-сервис: **хранение снапшотов миров** (opaque `bytes`, например JSON Ark). Не проксируется через gateway.

## Контракт

Полное описание API, ошибок, авторизации и правил полей: **[`CONTRACT.md`](CONTRACT.md)**.  
Нормативная схема сообщений: [`proto/world/v1/world.proto`](proto/world/v1/world.proto).

## Авторизация

Если в конфиге задан `service_token`, клиент передаёт:

- metadata `x-service-token: <token>`, или  
- `authorization: Bearer <token>`

Пустой `service_token` — без проверки (только для доверенной сети).

## Конфиг

`deploy/docker/config-world.toml` (Docker) или свой TOML:

- `bind_addr`, `database_url`, `log_level`, `service_token`

## Генерация proto

Из корня репозитория:

```bash
make gen-world
```

## Миграции

`migrations/` — Postgres, БД `world_db` (см. `docker-compose`). В т.ч. уникальность **непустого** `name` для upsert по имени (`GetWorldByName` + create/replace).

## Поиск по имени

RPC **`GetWorldByName`** — вернуть мир по полю `name` (после нормализации пробелов). Подробности в [`CONTRACT.md`](CONTRACT.md).

## Postman (gRPC)

Готовая коллекция и окружение: каталог [`postman/`](postman/) — см. [`postman/README.md`](postman/README.md).
