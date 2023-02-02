CREATE TYPE ITEM_TYPE AS ENUM (
    'image',
    'gif',
    'video',
    'video_note',
    'sticker',
    'voice',
    'audio'
);

CREATE TABLE Item (
    id SERIAL PRIMARY KEY,
    uid BIGINT NOT NULL,
    type ITEM_TYPE NOT NULL,
    alias VARCHAR(128) NOT NULL,
    file_id VARCHAR(128) NOT NULL,
    file_unique_id VARCHAR(32) NOT NULL,

    CONSTRAINT uk_item UNIQUE (uid, alias, file_unique_id)
);

CREATE INDEX IF NOT EXISTS idx_item ON Item(uid, alias);
