DO $$ BEGIN
    CREATE TYPE lang AS ENUM (
        'en',
        'ru'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

CREATE TABLE Users (
    uid bigint PRIMARY KEY,
    language lang
);
