## grpc_auth

Учебный микросервис аутентификации и авторизации на Go + gRPC.

### gRPC контракт (`auth-service/proto/auth/v1/auth.proto`)

- **Service `AuthService`**:
  - `Register(RegisterRequest) -> RegisterResponse` — регистрация пользователя по email/паролю, на выходе `user_id`.
  - `Login(LoginRequest) -> LoginResponse` — логин по email/паролю + `app_id`, возвращает пару токенов: `access_token` + `refresh_token` и их сроки жизни.
  - `IsAdmin(IsAdminRequest) -> IsAdminResponse` — проверка, является ли пользователь админом по `user_id`.
  - `Logout(LogoutRequest) -> LogoutResponse` — логаут по `refresh_token` (гасим конкретную сессию).
  - `RefreshToken(RefreshTokenRequest) -> RefreshTokenResponse` — по `refresh_token` выдаёт новую пару `access_token` + `refresh_token` и обновлённые сроки жизни.

### Модель токенов

- **Access-токен**:
  - формат: JWT;
  - срок жизни: **15 минут**;
  - хранение: **не храним в БД**, валидируем по подписи и сроку действия.
- **Refresh-токен**:
  - формат: JWT или случайная строка (решу на этапе реализации);
  - срок жизни: **7 дней**;
  - хранение: в БД **рядом с данными сессии** (user_id, app_id, refresh_expires_at, статус и т.п.).
- **Логаут**:
  - по `refresh_token` находим сессию в БД и помечаем её невалидной/удаляем;
  - связанные access-токены “умирают” естественно через 15 минут, новые уже не выдаются (нет живого refresh).

### Ближайшие шаги

- **Стабилизировать контракт**:
  - поправить мелкие синтаксические ошибки в `.proto` (например, не забыть `;` после RPC);
  - проверить, что `protoc` успешно генерит код.
- **Конфигурация**:
  - описать в `config.toml` и Go-структуре: срок жизни access/refresh, секрет JWT, параметры PostgreSQL;
  - завести пакет `config` в `auth-service`.
- **Хранение в БД**:
  - спроектировать таблицу `users`;
  - спроектировать таблицу для сессий/refresh-токенов (user_id, app_id, refresh_token, refresh_expires_at, status и т.п.).
- **Код**:
  - сгенерировать Go-код из `.proto` (server/client);
  - создать `cmd/auth-service/main.go` со скелетом gRPC-сервера и заглушками методов `AuthService`.
