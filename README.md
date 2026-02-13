## grpc_auth

Учебный микросервис аутентификации и авторизации на Go + gRPC.

Проект собирается по мотивам `sso` из GolangLessons, с упором на чистую архитектуру: отдельные слои домена, use‑case, инфраструктуры и транспорта.

### Структура репозитория

- **`auth-service/cmd/auth-service`** — entrypoint сервиса:
  - парсит `-config-path`;
  - читает `config.toml` через `BurntSushi/toml`;
  - инициализирует logger на `slog`;
  - сейчас только открывает соединение с БД (результат не используется) и стартует gRPC‑сервер.
- **`auth-service/internal/app`** — композиция приложения:
  - здесь должен собираться граф зависимостей: `sql.DB` → репозитории → usecase → gRPC‑handler;
  - пока оставлены `TODO` на инициализацию хранилища и auth‑сервиса.
- **`auth-service/internal/app/grpc`** — обёртка над `grpc.Server`:
  - регистрирует `AuthService` хендлер;
  - умеет запускаться и останавливаться с логированием.
- **`auth-service/internal/domain`** — доменная модель:
  - `User` (id, email, passHash);
  - `Session` (id, userID, appID) — упрощённое представление сессии, детали (`refresh_token`, `status` и т.п.) лежат в БД.
- **`auth-service/internal/repository`** — интерфейсы хранилищ:
  - `UserRepository` (создание пользователя, поиск по email, проверка админа);
  - `SessionRepository` (пока только `SessionByID`).
- **`auth-service/internal/infrastucture/sqlstore`** — реализация репозиториев поверх PostgreSQL:
  - `NewDB` — создание и ping `*sql.DB`;
  - `UserRepository`, `SessionRepository` — заглушки, которые ещё нужно реализовать (сейчас только заглушки с `_ = ctx` и т.п.).
- **`auth-service/internal/grpc/auth`** — gRPC‑слой:
  - регистрация сервера;
  - методы `Register`, `Login`, `IsAdmin`, `Logout`, `RefreshToken` пока `panic("implement me")`.
- **`auth-service/internal/config`** — конфигурация и логгер:
  - `Config` с полями `DatabaseURL`, `BindAddr`, TTL токенов, `LogLevel`;
  - `NewLogger` — JSON‑логгер на `slog` c уровнем из конфига.
- **`auth-service/migrations`** — миграции PostgreSQL (через `migrate`):
  - таблицы `users` и `sessions` под refresh‑токены, индексы под частые запросы.
- **`auth-service/proto/auth/v1`** — gRPC‑контракт и сгенерированный код.
- **корень репо**:
  - `Makefile` с целями `build-auth`, `start-auth`, `lint-auth`, `migrate-up`, `gen-auth`, `tidy-auth`, `gofmt`;
  - `config.toml` / `.env` и их примеры;
  - `go.work` — workspace с модулем `auth-service`.

### gRPC контракт (`auth-service/proto/auth/v1/auth.proto`)

- **Service `AuthService`**:
  - `Register(RegisterRequest) -> RegisterResponse` — регистрация пользователя по email/паролю, на выходе `user_id`;
  - `Login(LoginRequest) -> LoginResponse` — логин по email/паролю + `app_id`, возвращает `access_token`, `refresh_token` и сроки жизни через `google.protobuf.Timestamp`;
  - `IsAdmin(IsAdminRequest) -> IsAdminResponse` — проверка, является ли пользователь админом по `user_id`;
  - `Logout(LogoutRequest) -> LogoutResponse` — логаут по `refresh_token` (гасим конкретную сессию);
  - `RefreshToken(RefreshTokenRequest) -> RefreshTokenResponse` — по `refresh_token` выдаёт новую пару токенов и обновлённые TTL.

### Модель токенов (задумка)

- **Access‑токен**:
  - формат: JWT;
  - срок жизни: ~15 минут;
  - хранение: **не храним в БД**, валидируем по подписи, сроку действия и, при необходимости, по `session_id` в payload.
- **Refresh‑токен**:
  - формат: JWT или случайная строка (решение можно зафиксировать позже);
  - срок жизни: ~7 дней;
  - хранение: в таблице `sessions` вместе с `user_id`, `app_id`, `refresh_expires_at`, `status`.
- **Логаут**:
  - по `refresh_token` находим сессию в БД и помечаем её невалидной/удаляем;
  - новые access‑токены больше не выдаём, старые доживают до своего TTL.

### Как запустить (черновик себе)

1. **Поднять PostgreSQL** (локально или через Docker).
2. **Прописать строку подключения** в `config.toml`:
   - поле `database_url` должно совпадать с DSN для `github.com/lib/pq`;
   - там же задать `bind_addr`, TTL токенов и `log_level`.
3. **Прогнать миграции** (утилита `migrate` должна быть установлена):
   - команда: `make migrate-up` (нужен `DB_DSN` в окружении).
4. **Собрать бинарь**:
   - команда: `make build-auth` (собирает `./auth-service/cmd/auth-service`).
5. **Запустить сервис**:
   - команда: `make start-auth` (ожидает `./auth-service.exe` в корне и `config.toml` рядом).

### Статус по бизнес‑логике

На данный момент:

- протобуф‑контракт и миграции готовы;
- доменные модели и интерфейсы репозиториев определены;
- gRPC‑сервер поднимается, но RPC‑методы не реализованы (везде `panic`);
- репозитории в `sqlstore` — заглушки без SQL;
- usecase‑слой для auth ещё не написан и не провязан с gRPC‑хендлерами;
- JWT и работа с токенами не реализованы.

### План себе по реализации

1. **Дописать usecase‑слой** (`internal/usecase/auth.go`):
   - сценарии: `Register`, `Login`, `IsAdmin`, `Logout`, `RefreshToken`;
   - продумать контракт usecase vs grpc‑слой (DTO → domain → DTO).
2. **Реализовать репозитории** в `sqlstore`:
   - `SaveUser`, `UserByEmail`, `IsAdmin` по схеме `users`;
   - операции сессий по `refresh_token`, `user_id`, `app_id`, `status`.
3. **Пробросить зависимости в `internal/app`**:
   - `NewDB` → репозитории → usecase → gRPC‑handler (в `serverAPI` держать интерфейс usecase).
4. **JWT и безопасность**:
   - добавить библиотеку для JWT;
   - зафиксировать формат payload (user_id, session_id, is_admin, exp и т.п.);
   - реализовать генерацию и валидацию токенов.
5. **Тесты**:
   - написать unit‑тесты на usecase (через `testify` и мок‑репозитории);
   - по желанию добавить интеграционные тесты с тестовой БД.

Эту README можно расширять по мере появления докера, метрик, логирования запросов, трейсинга и т.д.
