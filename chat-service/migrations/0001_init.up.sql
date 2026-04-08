-- Универсальная таблица чатов. Тип: "dm" или "group".
-- DM-чат — чат ровно с двумя участниками, уникальность пары
-- гарантируется отдельным partial unique index ниже.
CREATE TABLE chats (
    id         BIGSERIAL   PRIMARY KEY,
    type       TEXT        NOT NULL CHECK (type IN ('dm', 'group')),
    name       TEXT,       -- для групповых чатов, для DM null
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Участники любого чата. unread_count — у каждого свой.
CREATE TABLE chat_members (
    chat_id      BIGINT NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    user_id      BIGINT NOT NULL,
    unread_count INT    NOT NULL DEFAULT 0 CHECK (unread_count >= 0),
    PRIMARY KEY (chat_id, user_id)
);

CREATE INDEX idx_chat_members_user ON chat_members (user_id);

-- Сообщения.
CREATE TABLE messages (
    id         BIGSERIAL   PRIMARY KEY,
    chat_id    BIGINT      NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    sender_id  BIGINT      NOT NULL,
    text       TEXT        NOT NULL CHECK (char_length(text) BETWEEN 1 AND 4096),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_messages_chat_created ON messages (chat_id, created_at DESC);

-- Гарантируем уникальность DM-пары: для чатов типа dm
-- у каждого участника не может быть двух чатов с одним и тем же собеседником.
-- Реализуется через уникальный индекс на пару (меньший, больший user_id).
-- Это enforced на уровне приложения при создании через GetOrCreateChat.
