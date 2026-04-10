# Postman: world-service (gRPC)

## Импорт

1. **Postman** → **Import** → выберите `World-Service.postman_collection.json`.
2. (Опционально) Импортируйте `Local.postman_environment.json` и выберите окружение **World Service — local**.

## Версия Postman

Нужен **десктопный Postman** с поддержкой **gRPC** и **multi-protocol collections**. Если после импорта запросы выглядят как «битые» или нет типа gRPC:

- Обновите Postman до актуальной версии.
- Создайте вручную **New → gRPC request**, импортируйте `../proto/world/v1/world.proto` в **Service definition** и скопируйте URL / сообщения из JSON вручную.

## Перед запуском

1. Поднимите `world-service` (Docker или локально) и миграции `world_db`.
2. В коллекции или окружении задайте:
   - **`grpcHost`** — `localhost:50054` (или `host.docker.internal:50054` с хоста к контейнеру).
   - **`worldId`** — уникальный id мира; повторный **CreateWorld** с тем же id вернёт `AlreadyExists`.
   - **`serviceToken`** — если в `config-world.toml` задан `service_token`, вставьте тот же секрет; иначе оставьте пустым.

## Прогон

- По одному запросу (**Invoke**) или **Collection Runner** в порядке: Create → Get → Replace → List → Delete.
- В **ReplaceWorldSnapshot** сейчас **`expectedVersion: 0`** (без optimistic lock), чтобы не зависеть от типа подстановки переменных в JSON. Для проверки lock задайте вручную число из **GetWorld** (`version`).

## Поле `snapshot`

В protobuf JSON поле `bytes` — **base64**. В примере Replace используется `e30=` (это `{}`).

## Ограничения Postman

Экспорт/импорт gRPC-коллекций в Postman иногда ведёт себя капризно; если JSON не подошёл — опора на импорт **только `.proto`** и ручное сохранение запросов в коллекцию остаётся надёжным запасным путём.
