ALTER TABLE Items ALTER COLUMN alias TYPE bigint USING alias::bigint;
ALTER TABLE Items ALTER COLUMN text TYPE bigint USING text::bigint;
