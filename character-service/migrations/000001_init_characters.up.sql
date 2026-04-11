CREATE TABLE characters (
    id UUID PRIMARY KEY,
    owner_user_id BIGINT NOT NULL,
    display_name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    data BYTEA NOT NULL DEFAULT '',
    schema_version INT NOT NULL DEFAULT 1,
    version BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX characters_owner_display_name_uq ON characters (owner_user_id, display_name);

CREATE INDEX characters_owner_user_id_idx ON characters (owner_user_id);
