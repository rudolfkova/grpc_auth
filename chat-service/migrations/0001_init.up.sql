CREATE TABLE chats (
    id         BIGSERIAL    PRIMARY KEY,
    user1_id   BIGINT       NOT NULL,
    user2_id   BIGINT       NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT now(),

    -- Гарантируем что user1_id < user2_id при вставке
    -- чтобы не было дублей (1,2) и (2,1)
    CONSTRAINT chk_user_order CHECK (user1_id < user2_id),
    CONSTRAINT uq_chat_participants UNIQUE (user1_id, user2_id)
);

CREATE TABLE messages (
    id         BIGSERIAL    PRIMARY KEY,
    chat_id    BIGINT       NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    sender_id  BIGINT       NOT NULL,
    text       TEXT         NOT NULL CHECK (char_length(text) BETWEEN 1 AND 4096),
    created_at TIMESTAMPTZ  NOT NULL DEFAULT now()
);

-- Для GetMessages (история чата по убыванию времени)
CREATE INDEX idx_messages_chat_created
    ON messages (chat_id, created_at DESC);

-- Для GetUserChats (все чаты пользователя)
CREATE INDEX idx_chats_user1 ON chats (user1_id);
CREATE INDEX idx_chats_user2 ON chats (user2_id);

-- Счётчик непрочитанных: отдельная таблица чище чем колонка в chats,
-- потому что у каждого участника свой счётчик
CREATE TABLE chat_members (
    chat_id      BIGINT  NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    user_id      BIGINT  NOT NULL,
    unread_count INT     NOT NULL DEFAULT 0 CHECK (unread_count >= 0),
    PRIMARY KEY (chat_id, user_id)
);

CREATE INDEX idx_chat_members_user ON chat_members (user_id);