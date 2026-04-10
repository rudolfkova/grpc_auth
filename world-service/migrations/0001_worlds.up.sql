CREATE TABLE worlds (
    id              TEXT PRIMARY KEY,
    name            TEXT        NOT NULL DEFAULT '',
    description     TEXT        NOT NULL DEFAULT '',
    snapshot        BYTEA       NOT NULL DEFAULT '\x',
    schema_version  INT         NOT NULL DEFAULT 1,
    version         BIGINT      NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
