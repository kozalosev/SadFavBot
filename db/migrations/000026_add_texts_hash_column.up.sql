CREATE OR REPLACE FUNCTION texts_hash_func(input text)
    RETURNS bytea
    LANGUAGE PLPGSQL
AS $$
BEGIN
    RETURN sha256(convert_to(input, 'UTF8'));
END;
$$;

ALTER TABLE Texts ADD COLUMN IF NOT EXISTS hash bytea;
UPDATE Texts SET hash = texts_hash_func(text || entities) WHERE hash IS NULL;
ALTER TABLE Texts ALTER COLUMN hash SET NOT NULL;

ALTER TABLE Texts DROP CONSTRAINT IF EXISTS texts_text_entities_key;
ALTER TABLE Texts ADD CONSTRAINT uk_texts_hash UNIQUE (hash);

CREATE OR REPLACE FUNCTION calculate_hash_for_texts()
    RETURNS TRIGGER
    LANGUAGE PLPGSQL
AS $$
BEGIN
    IF NEW.hash IS NULL THEN
        NEW.hash = texts_hash_func(NEW.text || NEW.entities);
    END IF;
    RETURN NEW;
END;
$$;

CREATE OR REPLACE TRIGGER trg_calculate_hash_for_texts BEFORE INSERT ON Texts
    FOR EACH ROW EXECUTE FUNCTION calculate_hash_for_texts();
