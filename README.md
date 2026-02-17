## grpc_auth

Учебный микросервис аутентификации и авторизации на Go + gRPC.

Собран по мотивам SSO из GolangLessons. Чистая архитектура: domain → usecase → транспорт (gRPC); репозитории и провайдер токенов — интерфейсы, реализации в infrastructure/provider.

---

### Структура репозитория

- **`auth-service/cmd/auth-service`** — точка входа:
  - флаг `-config-path`, чтение `config.toml` (BurntSushi/toml);
  - логгер (slog), подключение к БД через `sqlstore.NewDB`;
  - создание `AuthUseCase` и передача в `app.New`, запуск gRPC.
  - **Важно:** в main пока не собран полный граф зависимостей (репозитории, TokenProvider, TTL, секрет) — usecase создаётся пустым; для рабочего запуска нужно провести DB → репозитории → провайдер токенов → usecase и передать в app.

- **`auth-service/internal/app`** — композиция приложения:
  - принимает логгер, порт и `*usecase.AuthUseCase`;
  - создаёт gRPC-обёртку и возвращает `App` с `GRPCServer`.
  - слой не знает про БД и sqlstore — зависимости передаются снаружи.

- **`auth-service/internal/app/grpc`** — обёртка над `grpc.Server`:
  - регистрирует `AuthService` (интерфейс `grpcauth.Auth`);
  - `Run` / `MustRun`, `Stop` с логированием.

- **`auth-service/internal/domain`** — доменные модели:
  - `User` (ID, Email, PassHash);
  - `Token` (AccessToken, RefreshToken — строки; AccessExpireAt, RefreshExpireAt — time.Time);
  - `Session` — сессия refresh-токена: ID, UserID, AppID, RefreshExpiresAt, Status.

- **`auth-service/internal/repository`** — интерфейсы хранилищ:
  - `UserRepository`: SaveUser, UserByEmail, IsAdmin; константа `ErrUserNotFound`;
  - `SessionRepository`: SessionByID, CreateSession, RevokeByRefreshToken, SessionByRefreshToken.

- **`auth-service/internal/infrastructure/sqlstore`** — реализация под PostgreSQL:
  - `NewDB(databaseURL)` — открытие и ping;
  - `UserRepository`: SaveUser, UserByEmail, IsAdmin под таблицу `users`;
  - `SessionRepository`: SessionByID, CreateSession, RevokeByRefreshToken, SessionByRefreshToken под таблицу `sessions`;
  - `TokenProvider` (JWT access + random refresh) лежит здесь же (`token.go`).

- **`auth-service/provider`** — интерфейс `TokenProvider`:
  - CreateAccessToken(userID, sessionID, appID, exp) → access JWT;
  - CreateRefreshToken() → непрозрачная строка для refresh.
  - Реализация сейчас в `auth-service/internal/infrastructure/sqlstore/token.go`.

- **`auth-service/internal/grpc/auth`** — gRPC-слой:
  - интерфейс `Auth` с методами Register, Login, IsAdmin, Logout, RefreshToken (контракт usecase);
  - `serverAPI` вызывает usecase, маппит ответы в proto; все ошибки от usecase пока отдаются как `codes.Internal` — имеет смысл различать InvalidCredentials / InvalidRefreshToken → Unauthenticated.

- **`auth-service/internal/config`** — конфиг и логгер:
  - Config: DatabaseURL, BindAddr, AccessTokenTTL, RefreshTokenTTL (строки), LogLevel;
  - в `config.toml` используется `access_token_ttl` / `refresh_token_ttl`;
  - NewLogger — JSON slog по уровню из конфига.

- **`auth-service/migrations`** — миграции PostgreSQL (golang-migrate):
  - таблицы `users` (id, email, password_hash, is_admin, created_at, updated_at), `sessions` (id, user_id, app_id, refresh_token, refresh_expires_at, status, created_at, updated_at);
  - индексы по email, refresh_token, (user_id, app_id, status).

- **`auth-service/proto/auth/v1`** — gRPC-контракт и сгенерированный код.

- **Корень репо:** Makefile (build-auth, start-auth, lint-auth, migrate-up, gen-auth, tidy-auth, gofmt), config.toml / .env, go.work.

---

### gRPC-контракт (auth-service/proto/auth/v1)

- **AuthService:** Register, Login (access + refresh + expires), IsAdmin, Logout (по refresh_token), RefreshToken (новая пара токенов с ротацией refresh).

---

### Модель токенов (текущая реализация)

- **Access:** JWT, короткий TTL; в payload — user_id, session_id, app_id, exp. Не хранится в БД; для мгновенного логаута при проверке доступа нужно дополнительно проверять, что сессия в БД активна (status = 'active').
- **Refresh:** непрозрачная строка (crypto/rand + base64), хранится в `sessions.refresh_token`, TTL в конфиге. При RefreshToken — ротация: старый отзывается, создаётся новая сессия с новым refresh.
- **Logout:** по refresh_token сессия помечается revoked; доступ по соответствующему access прекращается при проверке сессии в БД.

---

### Как запустить

1. Поднять PostgreSQL, прописать `database_url` в config.toml (формат DSN для lib/pq).
2. В config.toml задать `bind_addr`, `access_token_tll`, `refresh_token_ttl` (например "15m", "168h"), `log_level`.
3. Выполнить миграции: `make migrate-up` (нужен DB_DSN в окружении, если так заложено в Makefile).
4. Собрать: `make build-auth`.
5. Запустить: `make start-auth` (бинарь и config.toml в ожидаемых путях по Makefile).

**Сейчас main не собирает репозитории, провайдер токенов и конфиг usecase** — для реального запуска нужно в main (или в app, если решишь собирать граф там) создать DB, репозитории, TokenProvider, распарсить TTL и секрет, вызвать `usecase.NewAuthUseCase(...)` и передать в `app.New`.

---

### Статус и что доделать

- **Сделано:** домен, интерфейсы репозиториев и TokenProvider, полный usecase (Register, Login, IsAdmin, Logout, RefreshToken с ротацией), gRPC-хендлеры, миграции, конфиг и логгер. Архитектура слоёв выдержана: usecase не зависит от grpc и sqlstore.
- **Не доделано:**  
  - main не собирает зависимости (пустой AuthUseCase);  
  - маппинг pg-ошибок на доменные (`ErrUserAlreadyExists` и т.п.) можно сделать аккуратнее;  
  - в usecase остался импорт `google.golang.org/grpc/status/codes` — один раз возвращается status.Error (Login при ошибке не ErrUserNotFound); usecase лучше возвращать только обычные error;  
  - в gRPC-хендлерах все ошибки → Internal; добавить маппинг ErrInvalidCredentials / ErrInvalidRefreshToken в Unauthenticated;  
  - интеграционные тесты usecase есть под билд-тегом `integration` (см. `auth-service/internal/usecase/auth_integration_test.go`); юнит-тесты usecase на моках — в `auth-service/internal/usecase/auth_test.go`.

---

### Интеграционные тесты

- Файл: `auth-service/internal/usecase/auth_integration_test.go`
- Запуск (из корня репо):

```bash
go test ./auth-service/internal/usecase -tags=integration
```

- Используется `test_database_url` из `config.toml`. Перед запуском миграции должны быть применены к этой базе.

---
