DROP TRIGGER IF EXISTS trg_calculate_hash_for_texts ON Texts;
DROP FUNCTION IF EXISTS calculate_hash_for_texts;
DROP FUNCTION IF EXISTS texts_hash_func;

ALTER TABLE Texts DROP CONSTRAINT IF EXISTS uk_texts_hash;
ALTER TABLE Texts ADD CONSTRAINT texts_text_entities_key UNIQUE (text, entities);

ALTER TABLE Texts DROP COLUMN IF EXISTS hash;
