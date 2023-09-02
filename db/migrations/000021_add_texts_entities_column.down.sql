do $$
    DECLARE
        txt_id Texts.id%TYPE;
    BEGIN
        SELECT id INTO txt_id FROM Texts WHERE entities IS NOT NULL;
        DELETE FROM Favs f WHERE f.text_id = txt_id;
        DELETE FROM Texts t WHERE t.id = txt_id;
    END
$$;

ALTER TABLE Texts DROP CONSTRAINT IF EXISTS texts_text_entities_key;
ALTER TABLE Texts ADD CONSTRAINT texts_text_key UNIQUE (text);

ALTER TABLE Texts DROP COLUMN IF EXISTS entities;
