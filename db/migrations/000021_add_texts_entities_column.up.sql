ALTER TABLE Texts ADD COLUMN IF NOT EXISTS entities jsonb;

ALTER TABLE Texts DROP CONSTRAINT IF EXISTS texts_text_key;
ALTER TABLE Texts ADD CONSTRAINT texts_text_entities_key UNIQUE (text, entities);
