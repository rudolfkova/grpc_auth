CREATE TABLE users (
    id             BIGSERIAL PRIMARY KEY,
    email          TEXT        NOT NULL UNIQUE,
    password_hash  TEXT        NOT NULL,
    is_admin       BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE sessions (
    id                 BIGSERIAL PRIMARY KEY, -- session_id для JWT
    user_id            BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    app_id             INT         NOT NULL,
    refresh_token      TEXT        NOT NULL,
    refresh_expires_at TIMESTAMPTZ NOT NULL,
    status             TEXT        NOT NULL DEFAULT 'active', -- active / revoked
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_users_email ON users (email);

CREATE INDEX idx_sessions_user_id ON sessions (user_id);
CREATE INDEX idx_sessions_refresh_token ON sessions (refresh_token);
CREATE INDEX idx_sessions_user_app_status ON sessions (user_id, app_id, status);