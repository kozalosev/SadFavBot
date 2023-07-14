ALTER INDEX favs_pkey RENAME TO items_pkey;
ALTER INDEX uk_favs_file RENAME TO uk_item_file;
ALTER INDEX uk_favs_text RENAME TO uk_item_text;
ALTER INDEX idx_favs_uid_alias RENAME TO idx_items_uid_alias;

ALTER TABLE Favs RENAME CONSTRAINT fk_favs_alias TO fk_item_alias;
ALTER TABLE Favs RENAME CONSTRAINT fk_favs_text TO fk_item_text;
ALTER TABLE Favs RENAME CONSTRAINT ck_favs_content TO ck_item_data;

ALTER TABLE Favs RENAME COLUMN alias_id TO alias;
ALTER TABLE Favs RENAME COLUMN text_id TO text;
ALTER TABLE Favs RENAME TO Items;

CREATE OR REPLACE FUNCTION ensure_link_not_existing_alias()
    RETURNS TRIGGER
    LANGUAGE PLPGSQL
AS $$
BEGIN
    IF exists(SELECT id FROM items WHERE uid = NEW.uid AND alias = NEW.alias_id) THEN
        RAISE EXCEPTION 'Insertion of the link with the same name as an already existing fav is forbidden';
    END IF;
    RETURN NEW;
END;
$$;
