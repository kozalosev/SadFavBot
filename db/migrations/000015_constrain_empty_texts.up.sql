BEGIN;
    DELETE FROM Items WHERE text = (SELECT id FROM Texts WHERE length(text) = 0);
    DELETE FROM Texts WHERE length(text) = 0;
COMMIT;

ALTER TABLE Texts ADD CONSTRAINT texts_non_empty CHECK ( length(text) > 0 );
