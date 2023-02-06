CREATE TYPE lang AS ENUM (
    'en',
    'ru'
);

CREATE TABLE Users (
    uid bigint PRIMARY KEY,
    language lang
);
