CREATE TYPE item_type AS ENUM (
    'text',
    'image',
    'gif',
    'video',
    'video_note',
    'sticker',
    'voice',
    'audio'
);

CREATE TABLE items (
    id serial PRIMARY KEY,
    uid bigint NOT NULL,
    type item_type NOT NULL,
    alias varchar(128) NOT NULL,

    file_id varchar(128),
    file_unique_id varchar(32),
    text varchar(10000),

    CONSTRAINT uk_item_file UNIQUE (uid, alias, file_unique_id),
    CONSTRAINT uk_item_text UNIQUE (uid, alias, text),
    CONSTRAINT ck_item_data CHECK ( file_id IS NOT NULL AND file_unique_id IS NOT NULL OR text IS NOT NULL )
);

CREATE INDEX IF NOT EXISTS idx_text_items ON items(uid, alias);
