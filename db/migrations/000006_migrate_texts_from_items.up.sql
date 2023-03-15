BEGIN TRANSACTION ;
    INSERT INTO Texts (text)
    SELECT DISTINCT text FROM Items WHERE text IS NOT NULL;

    UPDATE Items i
    SET text = (
        SELECT t.id
        FROM Texts AS t
        WHERE t.text = i.text
    ) WHERE text IS NOT NULL;
COMMIT ;

ALTER TABLE Items ALTER COLUMN text TYPE bigint USING text::bigint;
ALTER TABLE Items ADD CONSTRAINT fk_item_text FOREIGN KEY (text) REFERENCES Texts(id);
