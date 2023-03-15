ALTER TABLE Items DROP CONSTRAINT IF EXISTS fk_item_text;
ALTER TABLE Items ALTER COLUMN text TYPE varchar(10000) USING text::varchar(10000);

BEGIN TRANSACTION ;
    UPDATE Items i
    SET text = (
        SELECT t.text
        FROM Texts AS t
        WHERE t.id = i.text::bigint
    ) WHERE text IS NOT NULL;

    -- noinspection SqlWithoutWhere
    DELETE FROM Texts;
COMMIT ;
