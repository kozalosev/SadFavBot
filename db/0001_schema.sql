CREATE TYPE ITEM_TYPE AS ENUM (
    'image',
    'gif',
    'video',
    'video_note',
    'sticker',
    'voice',
    'audio'
);

CREATE TYPE LANG AS ENUM (
    'en',
    'ru'
);

CREATE TABLE Item (
    id SERIAL PRIMARY KEY,
    uid BIGINT NOT NULL,
    type ITEM_TYPE NOT NULL,
    alias VARCHAR NOT NULL,
    file_id VARCHAR NOT NULL,

    CONSTRAINT uk_item UNIQUE (uid, alias, file_id)
);

CREATE INDEX IF NOT EXISTS idx_item ON Item(uid, alias);
