CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX IF NOT EXISTS idx_name_trg_key ON Aliases USING gin (name gin_trgm_ops);
