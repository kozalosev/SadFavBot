ALTER TABLE Items DROP CONSTRAINT IF EXISTS fk_item_alias;
ALTER TABLE Items ALTER COLUMN alias TYPE varchar(128) USING alias::varchar(128);

UPDATE Items
SET alias = (
    SELECT a.name
    FROM Aliases AS a
    WHERE a.id = alias::bigint
);

-- noinspection SqlWithoutWhere
DELETE FROM Aliases;
