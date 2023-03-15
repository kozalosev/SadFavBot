INSERT INTO Aliases (name)
SELECT DISTINCT alias FROM Items;

UPDATE Items
SET alias = (
    SELECT a.id
    FROM Aliases AS a
    WHERE a.name = alias
);

ALTER TABLE Items ALTER COLUMN alias TYPE bigint USING alias::bigint;
ALTER TABLE Items ADD CONSTRAINT fk_item_alias FOREIGN KEY (alias) REFERENCES Aliases(id);
